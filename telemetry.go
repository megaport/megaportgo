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

// getProductTelemetry queries GET /v2/product/{productType}/{productUid}/telemetry.
func getProductTelemetry(ctx context.Context, c *Client, productType string, req *GetTelemetryRequest) (*ServiceTelemetryResponse, error) {
	if req == nil {
		return nil, ErrTelemetryRequestNil
	}
	if req.ProductUID == "" {
		return nil, ErrTelemetryProductUIDRequired
	}
	path := fmt.Sprintf("/v2/product/%s/%s/telemetry", productType, url.PathEscape(req.ProductUID))
	return fetchTelemetry(ctx, c, path, req.Types, req.From, req.To, req.Days)
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
		return nil, fmt.Errorf("parsing telemetry response from %s (trace ID %s): %w", path, resp.Header.Get(headerTraceId), err)
	}
	return telemetryResp, nil
}
