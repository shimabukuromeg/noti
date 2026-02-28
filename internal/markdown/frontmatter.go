package markdown

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents YAML frontmatter fields
type Frontmatter struct {
	Title     string   `yaml:"title,omitempty"`
	Slug      string   `yaml:"slug,omitempty"`
	Date      string   `yaml:"date,omitempty"`
	Tags      []string `yaml:"tags,omitempty"`
	Excerpt   string   `yaml:"excerpt,omitempty"`
	Published bool     `yaml:"published,omitempty"`
	NotionID  string   `yaml:"notion_id,omitempty"`
}

// ParseResult holds the parsed frontmatter and body
type ParseResult struct {
	Frontmatter    Frontmatter
	Body           string
	HasFrontmatter bool
}

// Parse parses a markdown file content into frontmatter and body.
// Frontmatter is delimited by "---" at start and end.
// If no frontmatter found, returns HasFrontmatter=false with entire content as Body.
func Parse(content string) (*ParseResult, error) {
	trimmed := strings.TrimLeft(content, " \t\r\n")

	// Check if content starts with "---"
	if !strings.HasPrefix(trimmed, "---") {
		return &ParseResult{
			Body:           content,
			HasFrontmatter: false,
		}, nil
	}

	// Find end of opening delimiter line
	afterOpening := trimmed[3:]
	newlineIdx := strings.Index(afterOpening, "\n")
	if newlineIdx == -1 {
		// Only "---" with nothing else
		return &ParseResult{
			Body:           content,
			HasFrontmatter: false,
		}, nil
	}

	// Everything after the first "---\n"
	rest := afterOpening[newlineIdx+1:]

	// Find the closing "---"
	closingIdx := strings.Index(rest, "---")
	if closingIdx == -1 {
		return &ParseResult{
			Body:           content,
			HasFrontmatter: false,
		}, nil
	}

	// Ensure the closing "---" is at the beginning of a line
	if closingIdx > 0 && rest[closingIdx-1] != '\n' {
		return &ParseResult{
			Body:           content,
			HasFrontmatter: false,
		}, nil
	}

	yamlContent := rest[:closingIdx]
	body := rest[closingIdx+3:]

	// Skip the newline after the closing delimiter, plus optional blank line
	body = strings.TrimLeft(body, "\n")

	var fm Frontmatter
	if strings.TrimSpace(yamlContent) != "" {
		if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
			return nil, fmt.Errorf("failed to parse frontmatter YAML: %w", err)
		}
	}

	return &ParseResult{
		Frontmatter:    fm,
		Body:           body,
		HasFrontmatter: true,
	}, nil
}

// Render produces a complete markdown string with YAML frontmatter + body.
// If frontmatter has no fields set, returns just the body.
func Render(fm Frontmatter, body string) string {
	// Check if frontmatter is empty (zero value)
	if fm.Title == "" && fm.Slug == "" && fm.Date == "" && len(fm.Tags) == 0 &&
		fm.Excerpt == "" && !fm.Published && fm.NotionID == "" {
		if body != "" && !strings.HasSuffix(body, "\n") {
			return body + "\n"
		}
		return body
	}

	yamlBytes, err := yaml.Marshal(&fm)
	if err != nil {
		// Should not happen with simple struct; return body only as fallback
		return body
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(yamlBytes)
	sb.WriteString("---\n\n")
	sb.WriteString(body)

	result := sb.String()
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

// UpdateNotionID reads a file, updates/adds notion_id in frontmatter, and writes it back.
// If file has no frontmatter, adds it.
func UpdateNotionID(filePath string, notionID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	result, err := Parse(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter in %s: %w", filePath, err)
	}

	fm := result.Frontmatter
	fm.NotionID = notionID

	output := Render(fm, result.Body)

	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}
