package teable

import (
	"context"
	"encoding/json"
	"fmt"
)

// Schema creation types

type CreateTableRequest struct {
	Name   string              `json:"name"`
	Fields []CreateFieldRequest `json:"fields,omitempty"`
}

type CreateTableResponse struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Fields []FieldResult `json:"fields,omitempty"`
}

type CreateFieldRequest struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Options interface{} `json:"options,omitempty"`
}

type FieldResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type LinkFieldOptions struct {
	Relationship  string `json:"relationship"`
	ForeignTableID string `json:"foreignTableId"`
	IsOneWay      bool   `json:"isOneWay,omitempty"`
}

type SelectFieldOptions struct {
	Choices []SelectChoice `json:"choices"`
}

type SelectChoice struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// CreateTable creates a new table in a base, optionally with inline fields.
func (c *Client) CreateTable(ctx context.Context, baseID string, req CreateTableRequest) (*CreateTableResponse, error) {
	path := fmt.Sprintf("/api/base/%s/table", baseID)
	data, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}
	var resp CreateTableResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode create table response: %w", err)
	}
	return &resp, nil
}

// CreateField creates a new field on an existing table.
func (c *Client) CreateField(ctx context.Context, tableID string, req CreateFieldRequest) (*FieldResult, error) {
	path := fmt.Sprintf("/api/table/%s/field", tableID)
	data, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}
	var resp FieldResult
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode create field response: %w", err)
	}
	return &resp, nil
}
