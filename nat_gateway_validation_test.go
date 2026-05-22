package megaport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"
)

// NATGatewayValidationTestSuite covers the matrix-aware validator and the
// reference-data cache wired onto NATGatewayServiceOp.
type NATGatewayValidationTestSuite struct {
	suite.Suite
	client *Client
	server *httptest.Server
	mux    *http.ServeMux
}

func TestNATGatewayValidationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(NATGatewayValidationTestSuite))
}

func (suite *NATGatewayValidationTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)
	suite.client = NewClient(nil, nil)
	u, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = u
}

func (suite *NATGatewayValidationTestSuite) TearDownTest() {
	suite.server.Close()
}

// matrixValidator returns the live NATGatewayService asserted to the optional
// NATGatewayMatrixValidator interface. Centralising the assertion keeps the
// individual test bodies tidy and satisfies forcetypeassert.
func (suite *NATGatewayValidationTestSuite) matrixValidator() NATGatewayMatrixValidator {
	mv, ok := suite.client.NATGatewayService.(NATGatewayMatrixValidator)
	suite.Require().True(ok, "NATGatewayService should implement NATGatewayMatrixValidator")
	return mv
}

const validationMatrixJSON = `{
	"message": "Success",
	"terms": "https://www.megaport.com/legal/acceptable-use-policy",
	"data": [
		{"sessionCount": [1000, 2000, 4000], "speedMbps": 100},
		{"sessionCount": [8000, 16000], "speedMbps": 1000}
	]
}`

func (suite *NATGatewayValidationTestSuite) registerMatrix(counter *atomic.Int32) {
	suite.mux.HandleFunc("/v3/products/nat_gateways/sessions", func(w http.ResponseWriter, r *http.Request) {
		if counter != nil {
			counter.Add(1)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, validationMatrixJSON)
	})
}

func (suite *NATGatewayValidationTestSuite) TestValidateSpeedSession_Valid() {
	suite.registerMatrix(nil)
	suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(context.Background(), 100, 2000))
	suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(context.Background(), 1000, 16000))
}

func (suite *NATGatewayValidationTestSuite) TestValidateSpeedSession_UnsupportedSpeed() {
	suite.registerMatrix(nil)
	err := suite.matrixValidator().ValidateNATGatewaySpeedSession(context.Background(), 500, 4000)
	suite.True(errors.Is(err, ErrNATGatewaySpeedNotSupported), "expected ErrNATGatewaySpeedNotSupported, got %v", err)
	suite.Contains(err.Error(), "500")
	suite.Contains(err.Error(), "100")
	suite.Contains(err.Error(), "1000")
}

func (suite *NATGatewayValidationTestSuite) TestValidateSpeedSession_UnsupportedSessionCount() {
	suite.registerMatrix(nil)
	err := suite.matrixValidator().ValidateNATGatewaySpeedSession(context.Background(), 100, 9999)
	suite.True(errors.Is(err, ErrNATGatewaySessionCountNotSupported), "expected ErrNATGatewaySessionCountNotSupported, got %v", err)
	suite.Contains(err.Error(), "9999")
	suite.Contains(err.Error(), "100")
}

func (suite *NATGatewayValidationTestSuite) TestValidateSpeedSession_CachesMatrix() {
	var hits atomic.Int32
	suite.registerMatrix(&hits)
	ctx := context.Background()

	for i := 0; i < 4; i++ {
		suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(ctx, 100, 1000))
	}
	suite.Equal(int32(1), hits.Load(), "matrix should be fetched once and cached")
}

func (suite *NATGatewayValidationTestSuite) TestValidateSpeedSession_PropagatesFetchError() {
	suite.mux.HandleFunc("/v3/products/nat_gateways/sessions", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"boom"}`, http.StatusInternalServerError)
	})
	err := suite.matrixValidator().ValidateNATGatewaySpeedSession(context.Background(), 100, 1000)
	suite.Error(err)
	suite.False(errors.Is(err, ErrNATGatewaySpeedNotSupported))
	suite.False(errors.Is(err, ErrNATGatewaySessionCountNotSupported))
}

func (suite *NATGatewayValidationTestSuite) TestValidateSpeedSession_CacheInvalidatedOn401() {
	var matrixHits atomic.Int32
	suite.registerMatrix(&matrixHits)
	// Separate endpoint we can hit to trigger the auth-failure invalidation.
	suite.mux.HandleFunc("/v3/products/nat_gateways/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	})
	natSvc := suite.client.NATGatewayService
	ctx := context.Background()

	// Prime the cache.
	suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(ctx, 100, 1000))
	suite.Equal(int32(1), matrixHits.Load())

	// Trigger a 401 from any other call. The Client should invalidate
	// registered reference-data caches.
	_, err := natSvc.GetNATGateway(ctx, "any-uid")
	suite.Error(err)

	// Next validation should re-fetch the matrix.
	suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(ctx, 100, 1000))
	suite.Equal(int32(2), matrixHits.Load(), "matrix cache should have been invalidated by 401")
}

func (suite *NATGatewayValidationTestSuite) TestValidateSpeedSession_CacheInvalidatedOn403() {
	var matrixHits atomic.Int32
	suite.registerMatrix(&matrixHits)
	suite.mux.HandleFunc("/v3/products/nat_gateways/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"forbidden"}`, http.StatusForbidden)
	})
	natSvc := suite.client.NATGatewayService
	ctx := context.Background()

	suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(ctx, 100, 1000))
	suite.Equal(int32(1), matrixHits.Load())

	_, err := natSvc.GetNATGateway(ctx, "any-uid")
	suite.Error(err)

	suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(ctx, 100, 1000))
	suite.Equal(int32(2), matrixHits.Load(), "matrix cache should have been invalidated by 403")
}

func (suite *NATGatewayValidationTestSuite) TestValidateSpeedSession_CacheNotInvalidatedOnNon401() {
	var matrixHits atomic.Int32
	suite.registerMatrix(&matrixHits)
	suite.mux.HandleFunc("/v3/products/nat_gateways/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"oops"}`, http.StatusInternalServerError)
	})
	natSvc := suite.client.NATGatewayService
	ctx := context.Background()

	suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(ctx, 100, 1000))
	suite.Equal(int32(1), matrixHits.Load())

	_, err := natSvc.GetNATGateway(ctx, "any-uid")
	suite.Error(err)

	suite.NoError(suite.matrixValidator().ValidateNATGatewaySpeedSession(ctx, 100, 1000))
	suite.Equal(int32(1), matrixHits.Load(), "matrix cache should survive non-auth errors")
}

// TestValidateNATGatewaySpeedSession_LazyCacheOnDirectInstantiation ensures
// that a NATGatewayServiceOp constructed directly (without
// NewNATGatewayService) does not panic on the first validator call: the
// session-matrix cache must be lazily initialized.
func TestValidateNATGatewaySpeedSession_LazyCacheOnDirectInstantiation(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/v3/products/nat_gateways/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, validationMatrixJSON)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(nil, nil)
	u, _ := url.Parse(server.URL)
	client.BaseURL = u

	// Direct struct literal, bypassing NewNATGatewayService entirely.
	svc := &NATGatewayServiceOp{Client: client}
	if err := svc.ValidateNATGatewaySpeedSession(context.Background(), 100, 1000); err != nil {
		t.Fatalf("unexpected error from lazily-initialized cache: %v", err)
	}
}

// Sanity check the shared structural validator behind both create and update.
func TestValidateNATGatewayCommonFields(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		productName string
		locationID  int
		speed       int
		term        int
		wantErr     error
	}{
		{"valid", "ng", 1, 500, 12, nil},
		{"missing name", "", 1, 500, 12, ErrNATGatewayProductNameRequired},
		{"bad location", "ng", 0, 500, 12, ErrNATGatewayLocationIDRequired},
		{"bad speed", "ng", 1, 0, 12, ErrNATGatewaySpeedRequired},
		{"bad term", "ng", 1, 500, 7, ErrNATGatewayInvalidTerm},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateNATGatewayCommonFields(tc.productName, tc.locationID, tc.speed, tc.term)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
