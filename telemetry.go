package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// telemetryValidationErrors carries a service's sentinel errors so the shared
// telemetry validation reports failures as that service's public errors.
type telemetryValidationErrors struct {
	productUIDRequired error
	typesRequired      error
	timeExclusive      error
	daysOutOfRange     error
	fromToIncomplete   error
	fromAfterTo        error
	rangeTooLong       error
}

// maxTelemetryRange is the API's maximum from/to time range (180 days).
const maxTelemetryRange = 180 * 24 * time.Hour

func validateTelemetryRequest(productUID string, types []string, from, to *time.Time, days *int32, errs telemetryValidationErrors) error {
	if productUID == "" {
		return errs.productUIDRequired
	}
	if len(types) == 0 {
		return errs.typesRequired
	}
	if days != nil && (from != nil || to != nil) {
		return errs.timeExclusive
	}
	if days != nil && (*days < 1 || *days > 180) {
		return errs.daysOutOfRange
	}
	if (from != nil) != (to != nil) {
		return errs.fromToIncomplete
	}
	if from != nil {
		if from.After(*to) {
			return errs.fromAfterTo
		}
		if to.Sub(*from) > maxTelemetryRange {
			return errs.rangeTooLong
		}
	}
	return nil
}

// fetchTelemetry queries a product telemetry endpoint and decodes the response.
func fetchTelemetry(ctx context.Context, c *Client, path string, types []string, from, to *time.Time, days *int32) (*ServiceTelemetryResponse, error) {
	params := url.Values{}
	for _, t := range types {
		params.Add("type", t)
	}
	if from != nil {
		params.Set("from", strconv.FormatInt(from.UnixMilli(), 10))
	}
	if to != nil {
		params.Set("to", strconv.FormatInt(to.UnixMilli(), 10))
	}
	if days != nil {
		params.Set("days", strconv.FormatInt(int64(*days), 10))
	}

	clientReq, err := c.NewRequest(ctx, http.MethodGet, path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	resp, err := c.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	telemetryResp := &ServiceTelemetryResponse{}
	if err := json.Unmarshal(buf.Bytes(), telemetryResp); err != nil {
		return nil, fmt.Errorf("parsing telemetry response from %s: %w", path, err)
	}
	return telemetryResp, nil
}
