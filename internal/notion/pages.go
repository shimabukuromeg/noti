package notion

import (
	"context"
	"fmt"
	"net/http"
)

// CreatePage creates a new page in a Notion database.
// POST /v1/pages
func (c *Client) CreatePage(ctx context.Context, dbID string, props PageProperties) (*Page, error) {
	req := createPageRequest{
		Parent: Parent{
			Type:       "database_id",
			DatabaseID: dbID,
		},
		Properties: props.ToNotionProperties(),
	}
	var result Page
	if err := c.do(ctx, http.MethodPost, "/pages", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdatePageProperties updates the properties of a Notion page.
// PATCH /v1/pages/{page_id}
func (c *Client) UpdatePageProperties(ctx context.Context, pageID string, props PageProperties) error {
	path := fmt.Sprintf("/pages/%s", pageID)
	req := updatePageRequest{
		Properties: props.ToNotionProperties(),
	}
	return c.do(ctx, http.MethodPatch, path, req, nil)
}

// ArchivePage archives a Notion page by setting archived=true.
// PATCH /v1/pages/{page_id}
func (c *Client) ArchivePage(ctx context.Context, pageID string) error {
	path := fmt.Sprintf("/pages/%s", pageID)
	archived := true
	req := updatePageRequest{
		Archived: &archived,
	}
	return c.do(ctx, http.MethodPatch, path, req, nil)
}

// GetPage retrieves a Notion page by its ID.
// GET /v1/pages/{page_id}
func (c *Client) GetPage(ctx context.Context, pageID string) (*Page, error) {
	path := fmt.Sprintf("/pages/%s", pageID)
	var result Page
	if err := c.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
