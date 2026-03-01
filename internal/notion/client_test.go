package notion

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// helper: create a test server that records the request and responds with the given status and body.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := NewClientWithBaseURL("test-token-abc123", srv.URL)
	return srv, client
}

func TestAuthorizationHeader(t *testing.T) {
	var gotAuth string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"object":"page","id":"page-1"}`))
	})

	_, err := client.GetPage(context.Background(), "page-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer test-token-abc123" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer test-token-abc123")
	}
}

func TestNotionVersionHeader(t *testing.T) {
	var gotVersion string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotVersion = r.Header.Get("Notion-Version")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"object":"page","id":"page-1"}`))
	})

	_, err := client.GetPage(context.Background(), "page-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotVersion != apiVersion {
		t.Errorf("Notion-Version header = %q, want %q", gotVersion, apiVersion)
	}
}

func TestGetPage(t *testing.T) {
	pageJSON := `{
		"object": "page",
		"id": "page-abc-123",
		"created_time": "2025-01-01T00:00:00.000Z",
		"last_edited_time": "2025-01-02T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-2"},
		"parent": {"type": "database_id", "database_id": "db-1"},
		"archived": false,
		"url": "https://www.notion.so/page-abc-123",
		"properties": {
			"Page": {
				"id": "title",
				"type": "title",
				"title": [{"type": "text", "text": {"content": "Test Page"}, "plain_text": "Test Page"}]
			},
			"Slug": {
				"id": "slug-id",
				"type": "rich_text",
				"rich_text": [{"type": "text", "text": {"content": "test-page"}, "plain_text": "test-page"}]
			}
		}
	}`

	var gotMethod, gotPath string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(pageJSON))
	})

	page, err := client.GetPage(context.Background(), "page-abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/pages/page-abc-123" {
		t.Errorf("path = %q, want /pages/page-abc-123", gotPath)
	}
	if page.ID != "page-abc-123" {
		t.Errorf("page.ID = %q, want %q", page.ID, "page-abc-123")
	}
	if page.Object != "page" {
		t.Errorf("page.Object = %q, want %q", page.Object, "page")
	}
	if page.URL != "https://www.notion.so/page-abc-123" {
		t.Errorf("page.URL = %q, want %q", page.URL, "https://www.notion.so/page-abc-123")
	}
	if page.Archived {
		t.Error("page.Archived = true, want false")
	}
	if title := page.Title(); title != "Test Page" {
		t.Errorf("page.Title() = %q, want %q", title, "Test Page")
	}
	if slug := page.Slug(); slug != "test-page" {
		t.Errorf("page.Slug() = %q, want %q", slug, "test-page")
	}
}

func TestGetPage_NotFound(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"object": "error",
			"status": 404,
			"code": "object_not_found",
			"message": "Could not find page with ID: nonexistent-id."
		}`))
	})

	_, err := client.GetPage(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	notionErr, ok := err.(*NotionError)
	if !ok {
		t.Fatalf("expected *NotionError, got %T: %v", err, err)
	}
	if notionErr.Status != 404 {
		t.Errorf("status = %d, want 404", notionErr.Status)
	}
	if notionErr.Code != "object_not_found" {
		t.Errorf("code = %q, want %q", notionErr.Code, "object_not_found")
	}
}

func TestCreatePage(t *testing.T) {
	responseJSON := `{
		"object": "page",
		"id": "new-page-id",
		"created_time": "2025-06-01T00:00:00.000Z",
		"last_edited_time": "2025-06-01T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "database_id", "database_id": "db-123"},
		"archived": false,
		"url": "https://www.notion.so/new-page-id",
		"properties": {
			"Page": {
				"id": "title",
				"type": "title",
				"title": [{"type": "text", "text": {"content": "New Blog Post"}, "plain_text": "New Blog Post"}]
			}
		}
	}`

	var gotMethod, gotPath string
	var gotBody createPageRequest
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	})

	props := PageProperties{
		Title:     "New Blog Post",
		Slug:      "new-blog-post",
		Date:      "2025-06-01",
		Tags:      []string{"go", "testing"},
		Published: true,
	}

	page, err := client.CreatePage(context.Background(), "db-123", props)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/pages" {
		t.Errorf("path = %q, want /pages", gotPath)
	}
	if gotBody.Parent.DatabaseID != "db-123" {
		t.Errorf("parent.database_id = %q, want %q", gotBody.Parent.DatabaseID, "db-123")
	}
	if gotBody.Parent.Type != "database_id" {
		t.Errorf("parent.type = %q, want %q", gotBody.Parent.Type, "database_id")
	}
	if page.ID != "new-page-id" {
		t.Errorf("page.ID = %q, want %q", page.ID, "new-page-id")
	}

	// Verify properties were sent in the request body
	if gotBody.Properties == nil {
		t.Fatal("expected properties in request body, got nil")
	}
	if _, ok := gotBody.Properties["Page"]; !ok {
		t.Error("expected 'Page' property in request body")
	}
	if _, ok := gotBody.Properties["Slug"]; !ok {
		t.Error("expected 'Slug' property in request body")
	}
	if _, ok := gotBody.Properties["Tags"]; !ok {
		t.Error("expected 'Tags' property in request body")
	}
	if _, ok := gotBody.Properties["Published"]; !ok {
		t.Error("expected 'Published' property in request body")
	}
}

func TestUpdatePageProperties(t *testing.T) {
	var gotMethod, gotPath string
	var gotBodyRaw map[string]json.RawMessage
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBodyRaw)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	props := PageProperties{
		Title:     "Updated Title",
		Published: false,
	}

	err := client.UpdatePageProperties(context.Background(), "page-to-update", props)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotPath != "/pages/page-to-update" {
		t.Errorf("path = %q, want /pages/page-to-update", gotPath)
	}
	if _, ok := gotBodyRaw["properties"]; !ok {
		t.Error("expected 'properties' key in request body")
	}
}

func TestArchivePage(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	err := client.ArchivePage(context.Background(), "page-to-archive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotPath != "/pages/page-to-archive" {
		t.Errorf("path = %q, want /pages/page-to-archive", gotPath)
	}
	archived, ok := gotBody["archived"]
	if !ok {
		t.Fatal("expected 'archived' key in request body")
	}
	if archived != true {
		t.Errorf("archived = %v, want true", archived)
	}
}

func TestRetrieveMarkdown(t *testing.T) {
	responseJSON := `{
		"object": "page_markdown",
		"id": "page-md-1",
		"markdown": "# Hello World\n\nThis is a test page.",
		"truncated": false,
		"unknown_block_ids": []
	}`

	var gotMethod, gotPath string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	})

	md, err := client.RetrieveMarkdown(context.Background(), "page-md-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/pages/page-md-1/markdown" {
		t.Errorf("path = %q, want /pages/page-md-1/markdown", gotPath)
	}
	if md.ID != "page-md-1" {
		t.Errorf("md.ID = %q, want %q", md.ID, "page-md-1")
	}
	if md.Markdown != "# Hello World\n\nThis is a test page." {
		t.Errorf("md.Markdown = %q, want %q", md.Markdown, "# Hello World\n\nThis is a test page.")
	}
	if md.Truncated {
		t.Error("md.Truncated = true, want false")
	}
	if md.Object != "page_markdown" {
		t.Errorf("md.Object = %q, want %q", md.Object, "page_markdown")
	}
}

func TestReplaceMarkdown(t *testing.T) {
	// ReplaceMarkdown: GET children → DELETE each block → PATCH insert_content
	blockChildren := `{
		"object": "list",
		"results": [
			{"object": "block", "id": "block-1"},
			{"object": "block", "id": "block-2"}
		],
		"has_more": false,
		"next_cursor": ""
	}`
	insertResponse := `{
		"object": "page_markdown",
		"id": "page-replace-1",
		"markdown": "# Replaced Content\n\nNew body here.",
		"truncated": false,
		"unknown_block_ids": []
	}`

	var deletedIDs []string
	var gotInsertBody updateMarkdownRequest
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/children"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(blockChildren))
		case r.Method == http.MethodDelete:
			// Record which block was deleted
			parts := strings.Split(r.URL.Path, "/")
			deletedIDs = append(deletedIDs, parts[len(parts)-1])
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodPatch:
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotInsertBody)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(insertResponse))
		}
	})

	md, err := client.ReplaceMarkdown(context.Background(), "page-replace-1", "# Replaced Content\n\nNew body here.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both blocks should have been deleted
	wantDeleted := []string{"block-1", "block-2"}
	if len(deletedIDs) != len(wantDeleted) {
		t.Fatalf("deleted block count = %d, want %d", len(deletedIDs), len(wantDeleted))
	}
	for i, id := range wantDeleted {
		if deletedIDs[i] != id {
			t.Errorf("deleted[%d] = %q, want %q", i, deletedIDs[i], id)
		}
	}

	// Final call should be insert_content
	if gotInsertBody.Type != "insert_content" {
		t.Errorf("insert type = %q, want %q", gotInsertBody.Type, "insert_content")
	}
	if gotInsertBody.InsertContent == nil {
		t.Fatal("expected insert_content in request body, got nil")
	}
	if gotInsertBody.InsertContent.Content != "# Replaced Content\n\nNew body here." {
		t.Errorf("content = %q, want %q", gotInsertBody.InsertContent.Content, "# Replaced Content\n\nNew body here.")
	}
	if md.Markdown != "# Replaced Content\n\nNew body here." {
		t.Errorf("md.Markdown = %q, want %q", md.Markdown, "# Replaced Content\n\nNew body here.")
	}
}

func TestInsertMarkdown(t *testing.T) {
	responseJSON := `{
		"object": "page_markdown",
		"id": "page-insert-1",
		"markdown": "# Existing\n\n## Appended Section",
		"truncated": false,
		"unknown_block_ids": []
	}`

	var gotMethod, gotPath string
	var gotBody updateMarkdownRequest
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	})

	md, err := client.InsertMarkdown(context.Background(), "page-insert-1", "## Appended Section")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotPath != "/pages/page-insert-1/markdown" {
		t.Errorf("path = %q, want /pages/page-insert-1/markdown", gotPath)
	}
	if gotBody.Type != "insert_content" {
		t.Errorf("type = %q, want %q", gotBody.Type, "insert_content")
	}
	if gotBody.InsertContent == nil {
		t.Fatal("expected insert_content in request body, got nil")
	}
	if gotBody.InsertContent.Content != "## Appended Section" {
		t.Errorf("content = %q, want %q", gotBody.InsertContent.Content, "## Appended Section")
	}
	if md.ID != "page-insert-1" {
		t.Errorf("md.ID = %q, want %q", md.ID, "page-insert-1")
	}
}

func TestQueryDatabase(t *testing.T) {
	responseJSON := `{
		"object": "list",
		"results": [
			{
				"object": "page",
				"id": "page-1",
				"created_time": "2025-06-01T00:00:00.000Z",
				"last_edited_time": "2025-06-01T00:00:00.000Z",
				"created_by": {"object": "user", "id": "u1"},
				"last_edited_by": {"object": "user", "id": "u1"},
				"parent": {"type": "database_id", "database_id": "db-query-1"},
				"archived": false,
				"url": "https://www.notion.so/page-1",
				"properties": {}
			},
			{
				"object": "page",
				"id": "page-2",
				"created_time": "2025-05-01T00:00:00.000Z",
				"last_edited_time": "2025-05-01T00:00:00.000Z",
				"created_by": {"object": "user", "id": "u1"},
				"last_edited_by": {"object": "user", "id": "u1"},
				"parent": {"type": "database_id", "database_id": "db-query-1"},
				"archived": false,
				"url": "https://www.notion.so/page-2",
				"properties": {}
			}
		],
		"has_more": true,
		"next_cursor": "cursor-abc"
	}`

	var gotMethod, gotPath string
	var gotBody queryDatabaseRequest
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	})

	published := true
	result, err := client.QueryDatabase(context.Background(), "db-query-1", QueryOptions{
		PageSize:  10,
		Published: &published,
		Tag:       "go",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/databases/db-query-1/query" {
		t.Errorf("path = %q, want /databases/db-query-1/query", gotPath)
	}
	if gotBody.PageSize != 10 {
		t.Errorf("page_size = %d, want 10", gotBody.PageSize)
	}

	// Verify sorts
	if len(gotBody.Sorts) != 1 {
		t.Fatalf("sorts length = %d, want 1", len(gotBody.Sorts))
	}
	if gotBody.Sorts[0]["property"] != "Date" {
		t.Errorf("sort property = %q, want %q", gotBody.Sorts[0]["property"], "Date")
	}
	if gotBody.Sorts[0]["direction"] != "descending" {
		t.Errorf("sort direction = %q, want %q", gotBody.Sorts[0]["direction"], "descending")
	}

	// Verify filter (compound "and" filter with Published + Tag)
	if gotBody.Filter == nil {
		t.Fatal("expected filter in request body, got nil")
	}
	andFilters, ok := gotBody.Filter["and"]
	if !ok {
		t.Fatal("expected 'and' compound filter")
	}
	filterSlice, ok := andFilters.([]interface{})
	if !ok {
		t.Fatalf("expected filter 'and' to be a slice, got %T", andFilters)
	}
	if len(filterSlice) != 2 {
		t.Errorf("filter 'and' length = %d, want 2", len(filterSlice))
	}

	// Verify result parsing
	if len(result.Results) != 2 {
		t.Fatalf("results length = %d, want 2", len(result.Results))
	}
	if result.Results[0].ID != "page-1" {
		t.Errorf("results[0].ID = %q, want %q", result.Results[0].ID, "page-1")
	}
	if result.Results[1].ID != "page-2" {
		t.Errorf("results[1].ID = %q, want %q", result.Results[1].ID, "page-2")
	}
	if !result.HasMore {
		t.Error("has_more = false, want true")
	}
	if result.NextCursor != "cursor-abc" {
		t.Errorf("next_cursor = %q, want %q", result.NextCursor, "cursor-abc")
	}
}

func TestRateLimitRetry(t *testing.T) {
	var callCount int64

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt64(&callCount, 1)
		if n <= 2 {
			// First two calls return 429
			w.Header().Set("Retry-After", "1")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{
				"object": "error",
				"status": 429,
				"code": "rate_limited",
				"message": "Rate limited"
			}`))
			return
		}
		// Third call succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"object": "page",
			"id": "page-retry-ok",
			"created_time": "2025-01-01T00:00:00.000Z",
			"last_edited_time": "2025-01-01T00:00:00.000Z",
			"created_by": {"object": "user", "id": "u1"},
			"last_edited_by": {"object": "user", "id": "u1"},
			"parent": {"type": "database_id", "database_id": "db-1"},
			"archived": false,
			"url": "https://www.notion.so/page-retry-ok",
			"properties": {}
		}`))
	})

	// Override the client's HTTP timeout to be generous for retry sleeps
	client.httpClient.Timeout = 30 * time.Second

	start := time.Now()
	page, err := client.GetPage(context.Background(), "page-retry-ok")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != "page-retry-ok" {
		t.Errorf("page.ID = %q, want %q", page.ID, "page-retry-ok")
	}

	finalCount := atomic.LoadInt64(&callCount)
	if finalCount != 3 {
		t.Errorf("request count = %d, want 3 (2 retries + 1 success)", finalCount)
	}

	// Verify that retries introduced some delay (exponential backoff: 2s + 4s = 6s minimum)
	if elapsed < 2*time.Second {
		t.Errorf("elapsed = %v, expected at least 2s for retry backoff", elapsed)
	}
}

func TestUnauthorized(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{
			"object": "error",
			"status": 401,
			"code": "unauthorized",
			"message": "API token is invalid."
		}`))
	})

	_, err := client.GetPage(context.Background(), "any-page")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	notionErr, ok := err.(*NotionError)
	if !ok {
		t.Fatalf("expected *NotionError, got %T: %v", err, err)
	}
	if notionErr.Status != 401 {
		t.Errorf("status = %d, want 401", notionErr.Status)
	}
	if notionErr.Code != "unauthorized" {
		t.Errorf("code = %q, want %q", notionErr.Code, "unauthorized")
	}
	if notionErr.Message != "API token is invalid." {
		t.Errorf("message = %q, want %q", notionErr.Message, "API token is invalid.")
	}

	// Verify Error() string formatting
	expected := "notion: unauthorized (401): API token is invalid."
	if notionErr.Error() != expected {
		t.Errorf("Error() = %q, want %q", notionErr.Error(), expected)
	}
}
