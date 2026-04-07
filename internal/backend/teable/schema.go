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

// CreateFieldRequest is used for inline table creation.
// notNull is NOT supported during inline creation — use UpdateField afterward.
type CreateFieldRequest struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Options interface{} `json:"options,omitempty"`
}

// StandaloneFieldRequest is used for POST /api/table/{tableId}/field.
// Supports notNull.
type StandaloneFieldRequest struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	NotNull bool        `json:"notNull,omitempty"`
	Options interface{} `json:"options,omitempty"`
}

// UpdateFieldRequest is used for PATCH on an existing field.
type UpdateFieldRequest struct {
	NotNull *bool `json:"notNull,omitempty"`
}

type FieldResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type LinkFieldOptions struct {
	Relationship   string `json:"relationship"`
	ForeignTableID string `json:"foreignTableId"`
	IsOneWay       bool   `json:"isOneWay,omitempty"`
}

type SelectFieldOptions struct {
	Choices []SelectChoice `json:"choices"`
}

type SelectChoice struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// ListFields returns all fields on a table.
func (c *Client) ListFields(ctx context.Context, tableID string) ([]FieldResult, error) {
	path := fmt.Sprintf("/api/table/%s/field", tableID)
	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var fields []FieldResult
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("decode field list: %w", err)
	}
	return fields, nil
}

// CreateTable creates a new table in a base with inline fields.
// notNull is not supported during inline creation.
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

// CreateField creates a new field on an existing table. Supports notNull.
func (c *Client) CreateField(ctx context.Context, tableID string, req StandaloneFieldRequest) (*FieldResult, error) {
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

// UpdateField updates properties of an existing field (e.g., setting notNull).
func (c *Client) UpdateField(ctx context.Context, tableID, fieldID string, req UpdateFieldRequest) error {
	path := fmt.Sprintf("/api/table/%s/field/%s", tableID, fieldID)
	_, err := c.doRequest(ctx, "PATCH", path, req)
	return err
}
