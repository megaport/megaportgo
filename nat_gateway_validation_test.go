package megaport

import (
	"errors"
	"testing"
)

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
