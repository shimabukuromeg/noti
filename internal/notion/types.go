package notion

import (
	"fmt"
	"time"
)

// Page represents a Notion page object.
type Page struct {
	Object         string              `json:"object"`
	ID             string              `json:"id"`
	CreatedTime    time.Time           `json:"created_time"`
	LastEditedTime time.Time           `json:"last_edited_time"`
	CreatedBy      User                `json:"created_by"`
	LastEditedBy   User                `json:"last_edited_by"`
	Parent         Parent              `json:"parent"`
	Archived       bool                `json:"archived"`
	Properties     map[string]Property `json:"properties"`
	URL            string              `json:"url"`
}

// User represents a Notion user reference.
type User struct {
	Object string `json:"object"`
	ID     string `json:"id"`
}

// Parent represents the parent of a Notion page.
type Parent struct {
	Type       string `json:"type"`
	DatabaseID string `json:"database_id,omitempty"`
	PageID     string `json:"page_id,omitempty"`
}

// Property represents a Notion page property value.
type Property struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Title       []RichText     `json:"title,omitempty"`
	RichText    []RichText     `json:"rich_text,omitempty"`
	Date        *DateValue     `json:"date,omitempty"`
	MultiSelect []SelectOption `json:"multi_select,omitempty"`
	Checkbox    *bool          `json:"checkbox,omitempty"`
}

// RichText represents a Notion rich text object.
type RichText struct {
	Type      string       `json:"type"`
	Text      *TextContent `json:"text,omitempty"`
	PlainText string       `json:"plain_text"`
}

// TextContent represents the text content within a RichText object.
type TextContent struct {
	Content string `json:"content"`
}

// DateValue represents a Notion date property value.
type DateValue struct {
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
}

// SelectOption represents a Notion select or multi_select option.
type SelectOption struct {
	Name string `json:"name"`
}

// PageMarkdown represents the response from the markdown endpoint.
type PageMarkdown struct {
	Object          string   `json:"object"`
	ID              string   `json:"id"`
	Markdown        string   `json:"markdown"`
	Truncated       bool     `json:"truncated"`
	UnknownBlockIDs []string `json:"unknown_block_ids"`
}

// PageProperties is used for creating/updating page properties.
type PageProperties struct {
	Title     string
	Slug      string
	Date      string
	Tags      []string
	Excerpt   string
	Published bool
}

// ToNotionProperties converts PageProperties to Notion API property format.
func (p PageProperties) ToNotionProperties() map[string]interface{} {
	props := make(map[string]interface{})

	if p.Title != "" {
		props["Page"] = map[string]interface{}{
			"title": []map[string]interface{}{
				{
					"text": map[string]interface{}{
						"content": p.Title,
					},
				},
			},
		}
	}

	if p.Slug != "" {
		props["Slug"] = map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"text": map[string]interface{}{
						"content": p.Slug,
					},
				},
			},
		}
	}

	if p.Date != "" {
		props["Date"] = map[string]interface{}{
			"date": map[string]interface{}{
				"start": p.Date,
			},
		}
	}

	if len(p.Tags) > 0 {
		tags := make([]map[string]interface{}, len(p.Tags))
		for i, tag := range p.Tags {
			tags[i] = map[string]interface{}{
				"name": tag,
			}
		}
		props["Tags"] = map[string]interface{}{
			"multi_select": tags,
		}
	}

	if p.Excerpt != "" {
		props["Excerpt"] = map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"text": map[string]interface{}{
						"content": p.Excerpt,
					},
				},
			},
		}
	}

	props["Published"] = map[string]interface{}{
		"checkbox": p.Published,
	}

	return props
}

// QueryOptions holds options for database queries.
type QueryOptions struct {
	PageSize    int
	StartCursor string
	Published   *bool  // filter by Published checkbox
	Tag         string // filter by tag
}

// QueryResult represents the response from a database query.
type QueryResult struct {
	Object     string `json:"object"`
	Results    []Page `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

// createPageRequest is the request body for POST /v1/pages.
type createPageRequest struct {
	Parent     Parent                 `json:"parent"`
	Properties map[string]interface{} `json:"properties"`
}

// updatePageRequest is the request body for PATCH /v1/pages/{id}.
type updatePageRequest struct {
	Properties map[string]interface{} `json:"properties,omitempty"`
	Archived   *bool                  `json:"archived,omitempty"`
}

// updateMarkdownRequest is the request body for PATCH /v1/pages/{id}/markdown.
type updateMarkdownRequest struct {
	Type           string          `json:"type"`
	InsertContent  *insertContent  `json:"insert_content,omitempty"`
	ReplaceContent *replaceContent `json:"replace_content_range,omitempty"`
}

type insertContent struct {
	Content string `json:"content"`
	After   string `json:"after,omitempty"`
}

type replaceContent struct {
	Content              string `json:"content"`
	ContentRange         string `json:"content_range"`
	AllowDeletingContent bool   `json:"allow_deleting_content,omitempty"`
}

// queryDatabaseRequest is the request body for POST /v1/databases/{id}/query.
type queryDatabaseRequest struct {
	PageSize    int                    `json:"page_size,omitempty"`
	StartCursor string                 `json:"start_cursor,omitempty"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
	Sorts       []map[string]string    `json:"sorts,omitempty"`
}

// NotionError represents an API error response.
type NotionError struct {
	Object  string `json:"object"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *NotionError) Error() string {
	return fmt.Sprintf("notion: %s (%d): %s", e.Code, e.Status, e.Message)
}

// Title extracts the page title from properties.
func (p *Page) Title() string {
	if prop, ok := p.Properties["Page"]; ok && len(prop.Title) > 0 {
		return prop.Title[0].PlainText
	}
	// Fallback: look for "Name" or "title" type property
	for _, prop := range p.Properties {
		if prop.Type == "title" && len(prop.Title) > 0 {
			return prop.Title[0].PlainText
		}
	}
	return ""
}

// Slug extracts the slug from properties.
func (p *Page) Slug() string {
	if prop, ok := p.Properties["Slug"]; ok && len(prop.RichText) > 0 {
		return prop.RichText[0].PlainText
	}
	return ""
}

// DateStr extracts the date string from properties.
func (p *Page) DateStr() string {
	if prop, ok := p.Properties["Date"]; ok && prop.Date != nil {
		return prop.Date.Start
	}
	return ""
}

// Tags extracts the tags from properties.
func (p *Page) Tags() []string {
	if prop, ok := p.Properties["Tags"]; ok {
		tags := make([]string, len(prop.MultiSelect))
		for i, s := range prop.MultiSelect {
			tags[i] = s.Name
		}
		return tags
	}
	return nil
}

// ExcerptStr extracts the excerpt from properties.
func (p *Page) ExcerptStr() string {
	if prop, ok := p.Properties["Excerpt"]; ok && len(prop.RichText) > 0 {
		return prop.RichText[0].PlainText
	}
	return ""
}

// IsPublished extracts the published status from properties.
func (p *Page) IsPublished() bool {
	if prop, ok := p.Properties["Published"]; ok && prop.Checkbox != nil {
		return *prop.Checkbox
	}
	return false
}
