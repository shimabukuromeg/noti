package notion

import (
	"context"
	"fmt"
	"net/http"
)

// RetrieveMarkdown fetches the markdown content of a Notion page.
// GET /v1/pages/{page_id}/markdown
func (c *Client) RetrieveMarkdown(ctx context.Context, pageID string) (*PageMarkdown, error) {
	path := fmt.Sprintf("/pages/%s/markdown", pageID)
	var result PageMarkdown
	if err := c.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReplaceMarkdown replaces all markdown content on a Notion page.
// It deletes all existing blocks via the Blocks API, then inserts the new content.
// This avoids the fragile content_range matching approach which can leave stale
// content behind when a page is truncated or when ranges fail to match.
func (c *Client) ReplaceMarkdown(ctx context.Context, pageID string, content string) (*PageMarkdown, error) {
	if err := c.ClearPageBlocks(ctx, pageID); err != nil {
		return nil, fmt.Errorf("failed to clear page blocks: %w", err)
	}
	return c.InsertMarkdown(ctx, pageID, content)
}

// InsertMarkdown appends markdown content to a Notion page.
// PATCH /v1/pages/{page_id}/markdown with type=insert_content.
func (c *Client) InsertMarkdown(ctx context.Context, pageID string, content string) (*PageMarkdown, error) {
	path := fmt.Sprintf("/pages/%s/markdown", pageID)
	req := updateMarkdownRequest{
		Type: "insert_content",
		InsertContent: &insertContent{
			Content: content,
		},
	}
	var result PageMarkdown
	if err := c.do(ctx, http.MethodPatch, path, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
