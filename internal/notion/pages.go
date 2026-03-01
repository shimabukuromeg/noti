package notion

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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

// getBlockChildren retrieves one page of block children.
// GET /v1/blocks/{block_id}/children?start_cursor=...
func (c *Client) getBlockChildren(ctx context.Context, blockID, cursor string) (*BlockChildrenResult, error) {
	path := fmt.Sprintf("/blocks/%s/children", blockID)
	if cursor != "" {
		path += "?start_cursor=" + url.QueryEscape(cursor)
	}
	var result BlockChildrenResult
	if err := c.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// deleteBlock archives (deletes) a single block.
// DELETE /v1/blocks/{block_id}
func (c *Client) deleteBlock(ctx context.Context, blockID string) error {
	path := fmt.Sprintf("/blocks/%s", blockID)
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// ClearPageBlocks deletes all top-level blocks from a page by iterating
// through paginated children and calling DELETE on each.
func (c *Client) ClearPageBlocks(ctx context.Context, pageID string) error {
	cursor := ""
	for {
		result, err := c.getBlockChildren(ctx, pageID, cursor)
		if err != nil {
			return fmt.Errorf("failed to list blocks: %w", err)
		}
		for _, block := range result.Results {
			if err := c.deleteBlock(ctx, block.ID); err != nil {
				return fmt.Errorf("failed to delete block %s: %w", block.ID, err)
			}
		}
		if !result.HasMore {
			break
		}
		cursor = result.NextCursor
	}
	return nil
}
