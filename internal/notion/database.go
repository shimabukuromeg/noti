package notion

import (
	"context"
	"fmt"
	"net/http"
)

// QueryDatabase queries a Notion database with optional filters and pagination.
// POST /v1/databases/{db_id}/query
// Results are sorted by the "Date" property in descending order.
func (c *Client) QueryDatabase(ctx context.Context, dbID string, opts QueryOptions) (*QueryResult, error) {
	path := fmt.Sprintf("/databases/%s/query", dbID)

	req := queryDatabaseRequest{
		Sorts: []map[string]string{
			{
				"property":  "Date",
				"direction": "descending",
			},
		},
	}

	if opts.PageSize > 0 {
		req.PageSize = opts.PageSize
	}

	if opts.StartCursor != "" {
		req.StartCursor = opts.StartCursor
	}

	// Build filter
	filters := []map[string]interface{}{}

	if opts.Published != nil {
		filters = append(filters, map[string]interface{}{
			"property": "Published",
			"checkbox": map[string]interface{}{
				"equals": *opts.Published,
			},
		})
	}

	if opts.Tag != "" {
		filters = append(filters, map[string]interface{}{
			"property": "Tags",
			"multi_select": map[string]interface{}{
				"contains": opts.Tag,
			},
		})
	}

	if len(filters) == 1 {
		req.Filter = filters[0]
	} else if len(filters) > 1 {
		req.Filter = map[string]interface{}{
			"and": filters,
		}
	}

	var result QueryResult
	if err := c.do(ctx, http.MethodPost, path, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
