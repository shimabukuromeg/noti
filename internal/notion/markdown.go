package notion

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
// It first retrieves current content to build the content_range selector,
// then uses replace_content_range to overwrite everything.
// If the page is empty, it falls back to insert_content.
func (c *Client) ReplaceMarkdown(ctx context.Context, pageID string, content string) (*PageMarkdown, error) {
	// Retrieve current content to build content_range
	current, err := c.RetrieveMarkdown(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve current markdown: %w", err)
	}

	trimmed := strings.TrimSpace(current.Markdown)
	if trimmed == "" {
		// Page is empty, use insert instead
		return c.InsertMarkdown(ctx, pageID, content)
	}

	// Build content_range from first and last lines
	contentRange := buildContentRange(trimmed)

	path := fmt.Sprintf("/pages/%s/markdown", pageID)
	req := updateMarkdownRequest{
		Type: "replace_content_range",
		ReplaceContent: &replaceContent{
			Content:              content,
			ContentRange:         contentRange,
			AllowDeletingContent: true,
		},
	}
	var result PageMarkdown
	if err := c.do(ctx, http.MethodPatch, path, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// buildContentRange creates an ellipsis selector "first line...last line"
// that matches from the beginning to the end of the content.
func buildContentRange(markdown string) string {
	lines := strings.Split(markdown, "\n")

	first := strings.TrimSpace(lines[0])
	last := strings.TrimSpace(lines[len(lines)-1])

	// Truncate to reasonable length for matching
	if len(first) > 80 {
		first = first[:80]
	}
	if len(last) > 80 {
		last = last[:80]
	}

	if first == last {
		return first
	}

	return first + "..." + last
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
