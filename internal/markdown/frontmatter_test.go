package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse_ValidFrontmatterAllFields(t *testing.T) {
	content := `---
title: My Post
slug: my-post
date: "2025-01-15"
tags:
  - go
  - cli
excerpt: A short excerpt
published: true
notion_id: abc123
---

# Hello World

This is the body.
`
	result, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Fatal("expected HasFrontmatter to be true")
	}
	if result.Frontmatter.Title != "My Post" {
		t.Errorf("expected title 'My Post', got %q", result.Frontmatter.Title)
	}
	if result.Frontmatter.Slug != "my-post" {
		t.Errorf("expected slug 'my-post', got %q", result.Frontmatter.Slug)
	}
	if result.Frontmatter.Date != "2025-01-15" {
		t.Errorf("expected date '2025-01-15', got %q", result.Frontmatter.Date)
	}
	if len(result.Frontmatter.Tags) != 2 || result.Frontmatter.Tags[0] != "go" || result.Frontmatter.Tags[1] != "cli" {
		t.Errorf("expected tags [go cli], got %v", result.Frontmatter.Tags)
	}
	if result.Frontmatter.Excerpt != "A short excerpt" {
		t.Errorf("expected excerpt 'A short excerpt', got %q", result.Frontmatter.Excerpt)
	}
	if !result.Frontmatter.Published {
		t.Error("expected published to be true")
	}
	if result.Frontmatter.NotionID != "abc123" {
		t.Errorf("expected notion_id 'abc123', got %q", result.Frontmatter.NotionID)
	}
	if !strings.Contains(result.Body, "# Hello World") {
		t.Errorf("expected body to contain '# Hello World', got %q", result.Body)
	}
	if !strings.Contains(result.Body, "This is the body.") {
		t.Errorf("expected body to contain 'This is the body.', got %q", result.Body)
	}
}

func TestParse_PartialFrontmatter(t *testing.T) {
	content := `---
title: Only Title
---

Body content here.
`
	result, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Fatal("expected HasFrontmatter to be true")
	}
	if result.Frontmatter.Title != "Only Title" {
		t.Errorf("expected title 'Only Title', got %q", result.Frontmatter.Title)
	}
	if result.Frontmatter.Slug != "" {
		t.Errorf("expected empty slug, got %q", result.Frontmatter.Slug)
	}
	if result.Frontmatter.Published {
		t.Error("expected published to be false")
	}
	if !strings.Contains(result.Body, "Body content here.") {
		t.Errorf("expected body to contain 'Body content here.', got %q", result.Body)
	}
}

func TestParse_NoFrontmatter(t *testing.T) {
	content := `# Just a heading

Some plain markdown without frontmatter.
`
	result, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasFrontmatter {
		t.Fatal("expected HasFrontmatter to be false")
	}
	if result.Body != content {
		t.Errorf("expected body to be entire content, got %q", result.Body)
	}
}

func TestParse_EmptyContent(t *testing.T) {
	result, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasFrontmatter {
		t.Fatal("expected HasFrontmatter to be false")
	}
	if result.Body != "" {
		t.Errorf("expected empty body, got %q", result.Body)
	}
}

func TestParse_FrontmatterOnly_NoBody(t *testing.T) {
	content := `---
title: No Body
slug: no-body
---
`
	result, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Fatal("expected HasFrontmatter to be true")
	}
	if result.Frontmatter.Title != "No Body" {
		t.Errorf("expected title 'No Body', got %q", result.Frontmatter.Title)
	}
	if result.Frontmatter.Slug != "no-body" {
		t.Errorf("expected slug 'no-body', got %q", result.Frontmatter.Slug)
	}
	if strings.TrimSpace(result.Body) != "" {
		t.Errorf("expected empty body, got %q", result.Body)
	}
}

func TestParse_LeadingWhitespace(t *testing.T) {
	content := `
---
title: Indented Start
---

Body after whitespace.
`
	result, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Fatal("expected HasFrontmatter to be true")
	}
	if result.Frontmatter.Title != "Indented Start" {
		t.Errorf("expected title 'Indented Start', got %q", result.Frontmatter.Title)
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	content := `---
title: [invalid yaml
  : broken
---

Body.
`
	_, err := Parse(content)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestRender_FullFrontmatter(t *testing.T) {
	fm := Frontmatter{
		Title:     "My Post",
		Slug:      "my-post",
		Date:      "2025-01-15",
		Tags:      []string{"go", "cli"},
		Excerpt:   "Short excerpt",
		Published: true,
		NotionID:  "abc123",
	}
	body := "# Hello\n\nWorld.\n"

	output := Render(fm, body)

	if !strings.HasPrefix(output, "---\n") {
		t.Error("expected output to start with '---\\n'")
	}
	if !strings.Contains(output, "title: My Post") {
		t.Error("expected output to contain 'title: My Post'")
	}
	if !strings.Contains(output, "slug: my-post") {
		t.Error("expected output to contain 'slug: my-post'")
	}
	if !strings.Contains(output, "notion_id: abc123") {
		t.Error("expected output to contain 'notion_id: abc123'")
	}
	if !strings.Contains(output, "published: true") {
		t.Error("expected output to contain 'published: true'")
	}
	if !strings.Contains(output, "---\n\n# Hello") {
		t.Error("expected closing delimiter followed by blank line and body")
	}
	if !strings.HasSuffix(output, "\n") {
		t.Error("expected output to end with newline")
	}
}

func TestRender_EmptyFrontmatter(t *testing.T) {
	fm := Frontmatter{}
	body := "Just some text.\n"

	output := Render(fm, body)

	if strings.Contains(output, "---") {
		t.Errorf("expected no frontmatter delimiters for empty frontmatter, got %q", output)
	}
	if output != "Just some text.\n" {
		t.Errorf("expected 'Just some text.\\n', got %q", output)
	}
}

func TestRender_EmptyFrontmatter_AddsTrailingNewline(t *testing.T) {
	fm := Frontmatter{}
	body := "No trailing newline"

	output := Render(fm, body)

	if !strings.HasSuffix(output, "\n") {
		t.Error("expected output to end with newline")
	}
}

func TestRender_OmitsEmptyFields(t *testing.T) {
	fm := Frontmatter{
		Title: "Only Title",
	}
	body := "Body.\n"

	output := Render(fm, body)

	if strings.Contains(output, "slug:") {
		t.Error("expected slug to be omitted")
	}
	if strings.Contains(output, "date:") {
		t.Error("expected date to be omitted")
	}
	if strings.Contains(output, "tags:") {
		t.Error("expected tags to be omitted")
	}
	if strings.Contains(output, "excerpt:") {
		t.Error("expected excerpt to be omitted")
	}
	if strings.Contains(output, "published:") {
		t.Error("expected published to be omitted (false is zero value)")
	}
	if strings.Contains(output, "notion_id:") {
		t.Error("expected notion_id to be omitted")
	}
	if !strings.Contains(output, "title: Only Title") {
		t.Error("expected title to be present")
	}
}

func TestRoundTrip_ParseThenRender(t *testing.T) {
	original := Frontmatter{
		Title:     "Round Trip",
		Slug:      "round-trip",
		Date:      "2025-06-01",
		Tags:      []string{"test"},
		Excerpt:   "Testing round trip",
		Published: true,
		NotionID:  "rt-123",
	}
	originalBody := "# Round Trip Test\n\nThis should survive a round trip.\n"

	rendered := Render(original, originalBody)

	parsed, err := Parse(rendered)
	if err != nil {
		t.Fatalf("unexpected error on round-trip parse: %v", err)
	}
	if !parsed.HasFrontmatter {
		t.Fatal("expected HasFrontmatter to be true after round trip")
	}
	if parsed.Frontmatter.Title != original.Title {
		t.Errorf("title mismatch: got %q, want %q", parsed.Frontmatter.Title, original.Title)
	}
	if parsed.Frontmatter.Slug != original.Slug {
		t.Errorf("slug mismatch: got %q, want %q", parsed.Frontmatter.Slug, original.Slug)
	}
	if parsed.Frontmatter.Date != original.Date {
		t.Errorf("date mismatch: got %q, want %q", parsed.Frontmatter.Date, original.Date)
	}
	if len(parsed.Frontmatter.Tags) != len(original.Tags) {
		t.Errorf("tags length mismatch: got %d, want %d", len(parsed.Frontmatter.Tags), len(original.Tags))
	}
	if parsed.Frontmatter.Excerpt != original.Excerpt {
		t.Errorf("excerpt mismatch: got %q, want %q", parsed.Frontmatter.Excerpt, original.Excerpt)
	}
	if parsed.Frontmatter.Published != original.Published {
		t.Errorf("published mismatch: got %v, want %v", parsed.Frontmatter.Published, original.Published)
	}
	if parsed.Frontmatter.NotionID != original.NotionID {
		t.Errorf("notion_id mismatch: got %q, want %q", parsed.Frontmatter.NotionID, original.NotionID)
	}
	if parsed.Body != originalBody {
		t.Errorf("body mismatch:\ngot:  %q\nwant: %q", parsed.Body, originalBody)
	}
}

func TestUpdateNotionID_WithExistingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")

	original := `---
title: Existing Post
slug: existing-post
---

# Content

Some body text.
`
	if err := os.WriteFile(filePath, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := UpdateNotionID(filePath, "notion-456"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "notion_id: notion-456") {
		t.Errorf("expected notion_id in output, got:\n%s", content)
	}
	if !strings.Contains(content, "title: Existing Post") {
		t.Errorf("expected title to be preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "slug: existing-post") {
		t.Errorf("expected slug to be preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "# Content") {
		t.Errorf("expected body to be preserved, got:\n%s", content)
	}
}

func TestUpdateNotionID_WithoutFrontmatter(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")

	original := `# No Frontmatter

Just plain markdown.
`
	if err := os.WriteFile(filePath, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := UpdateNotionID(filePath, "notion-789"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "notion_id: notion-789") {
		t.Errorf("expected notion_id in output, got:\n%s", content)
	}
	if !strings.Contains(content, "# No Frontmatter") {
		t.Errorf("expected original body to be preserved, got:\n%s", content)
	}
}

func TestUpdateNotionID_NonexistentFile(t *testing.T) {
	err := UpdateNotionID("/nonexistent/path/file.md", "id")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestUpdateNotionID_OverwritesExistingNotionID(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")

	original := `---
title: Post
notion_id: old-id
---

Body.
`
	if err := os.WriteFile(filePath, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := UpdateNotionID(filePath, "new-id"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "notion_id: new-id") {
		t.Errorf("expected new notion_id, got:\n%s", content)
	}
	if strings.Contains(content, "old-id") {
		t.Errorf("expected old notion_id to be replaced, got:\n%s", content)
	}
}
