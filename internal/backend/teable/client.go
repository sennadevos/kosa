package teable

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client is a low-level HTTP client for the Teable REST API.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		http:    &http.Client{},
	}
}

// RawRecord represents a record as returned by Teable's API.
type RawRecord struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

type listResponse struct {
	Records []RawRecord `json:"records"`
}

type createRequest struct {
	Records []createRecordEntry `json:"records"`
}

type createRecordEntry struct {
	Fields map[string]interface{} `json:"fields"`
}

type createResponse struct {
	Records []RawRecord `json:"records"`
}

type updateRequest struct {
	Record updateRecordEntry `json:"record"`
}

type updateRecordEntry struct {
	Fields map[string]interface{} `json:"fields"`
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("teable api %s %s: %d %s", method, path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) ListRecords(ctx context.Context, tableID string, filter string, sort string, limit int) ([]RawRecord, error) {
	params := url.Values{}
	if filter != "" {
		params.Set("filter", filter)
	}
	if sort != "" {
		params.Set("sort", sort)
	}
	if limit > 0 {
		params.Set("take", fmt.Sprintf("%d", limit))
	}

	path := fmt.Sprintf("/api/table/%s/record", tableID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp listResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}
	return resp.Records, nil
}

func (c *Client) GetRecord(ctx context.Context, tableID, recordID string) (*RawRecord, error) {
	path := fmt.Sprintf("/api/table/%s/record/%s", tableID, recordID)
	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var rec RawRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("decode record: %w", err)
	}
	return &rec, nil
}

func (c *Client) CreateRecord(ctx context.Context, tableID string, fields map[string]interface{}) (*RawRecord, error) {
	path := fmt.Sprintf("/api/table/%s/record", tableID)
	body := createRequest{
		Records: []createRecordEntry{{Fields: fields}},
	}

	data, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	var resp createResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode create response: %w", err)
	}
	if len(resp.Records) == 0 {
		return nil, fmt.Errorf("teable returned no records on create")
	}
	return &resp.Records[0], nil
}

func (c *Client) UpdateRecord(ctx context.Context, tableID, recordID string, fields map[string]interface{}) (*RawRecord, error) {
	path := fmt.Sprintf("/api/table/%s/record/%s", tableID, recordID)
	body := updateRequest{
		Record: updateRecordEntry{Fields: fields},
	}

	data, err := c.doRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	var rec RawRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("decode update response: %w", err)
	}
	return &rec, nil
}

func (c *Client) DeleteRecord(ctx context.Context, tableID, recordID string) error {
	path := fmt.Sprintf("/api/table/%s/record/%s", tableID, recordID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}
