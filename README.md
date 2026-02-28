# noti

CLI tool to sync local Markdown files with Notion pages using the [Notion Markdown API](https://developers.notion.com/reference/markdown).

## Install

```sh
go install github.com/shimabukuromeg/noti/cmd/noti@latest
```

Or download a prebuilt binary from [GitHub Releases](https://github.com/shimabukuromeg/noti/releases).

## Setup

Set the following environment variables:

```sh
export NOTION_TOKEN="your-notion-integration-token"
export NOTI_DATABASE_ID="your-database-id"
```

## Usage

### push - Push Markdown to Notion

```sh
noti push article.md                          # Update existing page (uses notion_id from frontmatter)
noti push article.md --database <DB_ID>       # Create new page in database
noti push article.md --page-id <PAGE_ID>      # Update specific page
noti push article.md --force                  # Skip conflict detection
```

### pull - Pull Notion page as Markdown

```sh
noti pull <page-id>                           # Output to stdout
noti pull <page-id> -o article.md             # Save to file
```

### list - List pages in database

```sh
noti list                                     # Default 20 pages
noti list --limit 50
noti list --json
noti list --published
noti list --tag Go
```

### delete - Archive a Notion page

```sh
noti delete <page-id>
noti delete <page-id> --force
```

### version

```sh
noti version
```

## Frontmatter

noti uses YAML frontmatter to map Markdown metadata to Notion database properties.

```yaml
---
title: "Article Title"
slug: my-article
date: 2026-03-01
tags:
  - Go
  - Notion
excerpt: "Brief description"
published: true
notion_id: "abc123-def456"  # Auto-added after push
---
```

| Frontmatter field | Notion property | Property type |
|---|---|---|
| `title` | Page | title |
| `slug` | Slug | rich_text |
| `date` | Date | date |
| `tags` | Tags | multi_select |
| `excerpt` | Excerpt | rich_text |
| `published` | Published | checkbox |

## Global Flags

- `--token`, `-t` -- Override `NOTION_TOKEN`
- `--database`, `-d` -- Override `NOTI_DATABASE_ID`

## License

MIT
