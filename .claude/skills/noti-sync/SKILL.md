---
name: noti-sync
description: |
  Sync local Markdown files with Notion using the noti CLI for blog writing workflows.
  Use when the user wants to: (1) push a local Markdown file to Notion, (2) pull a Notion page as Markdown,
  (3) list Notion database pages, (4) delete/archive a Notion page, (5) manage blog posts between local and Notion,
  (6) publish or update blog articles. Triggers on: "Notionに同期", "ブログを公開", "記事をpush", "Notionから取得",
  "noti push", "noti pull", "noti list", "記事一覧", "Notionに上げて", "Notionから引っ張って".
---

# noti — Markdown ↔ Notion Sync CLI

## Prerequisites

- `noti` binary installed (`go install github.com/shimabukuromeg/noti/cmd/noti@latest`)
- Authenticated via `noti login` (token at `~/.config/noti/token.json`) or `NOTION_TOKEN` env var
- `NOTI_DATABASE_ID` env var set to the target Notion database

Verify setup: `noti list --limit 1`

## Commands

```
noti push <file.md>                  # Create or update Notion page
noti push <file.md> --force          # Skip conflict detection
noti push <file.md> --database <ID>  # Specify database for new page
noti pull <page-id>                  # Output markdown to stdout
noti pull <page-id> -o <file.md>     # Save to file
noti list                            # List pages (default 20, date desc)
noti list --limit 50 --json          # JSON output
noti list --published --tag Go       # Filter
noti delete <page-id>                # Archive (with confirmation prompt)
noti delete <page-id> --force        # Archive without confirmation
```

## Frontmatter Format

```yaml
---
title: "記事タイトル"
slug: my-article
date: 2026-03-01
tags:
  - Go
  - Notion
excerpt: "記事の概要"
published: true
notion_id: "abc123-def456"  # Auto-added after first push
---
```

- `notion_id` is auto-written back to the file after `noti push` creates a new page.
- On subsequent pushes, `notion_id` determines which page to update.
- If frontmatter is absent, the filename (without extension) becomes the title.

## Workflow

### Push a new article

1. Confirm the markdown file path and content with user
2. Run `noti push <file.md>`
3. Verify `notion_id` was written back: check the frontmatter

### Update an existing article

1. Confirm the file has `notion_id` in frontmatter
2. Run `noti push <file.md>` (uses `notion_id` to find the page)
3. If conflict warning appears, ask user whether to `--force`

### Pull from Notion

1. Get the page ID (from `noti list` or frontmatter `notion_id`)
2. Run `noti pull <page-id> -o <file.md>` to save, or `noti pull <page-id>` for stdout

### List and find pages

1. Run `noti list` for overview
2. Use `--json` for programmatic access, `--published` / `--tag` for filtering

## Rules

- Always confirm with the user before running `noti push`, `noti delete`, or any write operation.
- When pushing, check if the file already has `notion_id` to determine create vs update.
- For `noti delete`, prefer using `--force` only when the user explicitly agrees.
- Display the Notion URL from push output so the user can verify in browser.
