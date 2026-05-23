package atlasclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/whatnick/daa_mcp/internal/model"
)

const DefaultBaseURL = "https://digital.atlas.gov.au/api/search/v1"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = DefaultBaseURL
	}

	return &Client{
		baseURL: strings.TrimRight(base, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) ListCollections(ctx context.Context) (*model.CollectionsResponse, error) {
	respBody, err := c.get(ctx, "/collections", nil)
	if err != nil {
		return nil, err
	}

	var out model.CollectionsResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("decode collections response: %w", err)
	}

	return &out, nil
}

func (c *Client) SearchItems(ctx context.Context, collectionID, q string, limit, startIndex int) (*model.ItemsResponse, error) {
	params := url.Values{}
	if q != "" {
		params.Set("q", q)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if startIndex > 0 {
		params.Set("startindex", strconv.Itoa(startIndex))
	}

	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))
	respBody, err := c.get(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var out model.ItemsResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("decode items response: %w", err)
	}
	out.Raw = append(out.Raw[:0], respBody...)

	return &out, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	endpoint := c.baseURL + path
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request %s: %w", endpoint, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response %s: %w", endpoint, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("upstream %s returned %d: %s", endpoint, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}
