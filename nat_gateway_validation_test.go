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

func TestNATGatewaySpeedSessionSupported(t *testing.T) {
	t.Parallel()
	// Synthetic matrix; values are illustrative, not real availability data.
	matrix := []*NATGatewaySession{
		{SpeedMbps: 1000, SessionCount: []int{16000, 32000}},
		nil,
		{SpeedMbps: 5000, SessionCount: []int{128000}},
	}
	cases := []struct {
		name           string
		speed, session int
		wantSupported  bool
		wantSpeedKnown bool
	}{
		{"valid pair", 1000, 32000, true, true},
		{"unsupported speed", 500, 16000, false, false},
		{"bad session at good speed", 1000, 99999, false, true},
		{"skips nil entries", 5000, 128000, true, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := NATGatewaySpeedSessionSupported(matrix, tc.speed, tc.session)
			if r.Supported != tc.wantSupported || r.SpeedSupported != tc.wantSpeedKnown {
				t.Fatalf("got supported=%v speedSupported=%v, want %v/%v",
					r.Supported, r.SpeedSupported, tc.wantSupported, tc.wantSpeedKnown)
			}
		})
	}
}
