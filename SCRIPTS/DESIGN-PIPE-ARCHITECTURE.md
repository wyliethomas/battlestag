# Pipe-Friendly Module Architecture

## Design Principle

**Modules output JSON to stdout for piping, human-readable logs to stderr**

```bash
# When run standalone (no pipe), shows human output
youtube_to_obsidian 'URL'
# Logs to stderr, saved to vault

# When piped, outputs JSON
youtube_to_obsidian 'URL' | push_to_rag
# JSON on stdout → push_to_rag
# Human logs to stderr (visible)
```

## Standard Output Format

All pipeable modules output JSON to stdout:

```json
{
  "status": "success",
  "module": "youtube_to_obsidian",
  "tier": 2,
  "data": {
    "file_path": "/path/to/note.md",
    "title": "Video Title",
    "content": "Full markdown content...",
    "metadata": {
      "video_id": "abc123",
      "channel": "Channel Name",
      "duration": "10:23",
      "word_count": 2131
    }
  },
  "cost": 0.08,
  "timestamp": "2025-12-23T00:19:58Z"
}
```

## Pipe Detection

Modules detect if they're being piped:

```bash
if [[ -t 1 ]]; then
    # stdout is a terminal (not piped)
    # Show human-friendly output
else
    # stdout is piped
    # Output JSON only
fi
```

## Example Workflows

### Simple Pipe
```bash
youtube_to_obsidian 'URL' | push_to_rag
```

### Multi-Stage Pipeline
```bash
youtube_to_obsidian 'URL' | \
  extract_key_points | \
  push_to_rag
```

### Conditional Processing
```bash
youtube_to_obsidian 'URL' | \
  jq 'select(.data.word_count > 1000)' | \
  push_to_rag
```

### Batch Processing
```bash
cat video_urls.txt | while read url; do
  youtube_to_obsidian "$url" | push_to_rag
done
```

## Module Categories

### Source Modules (Tier 1 & 2)
- `youtube_to_obsidian` - Extract video transcripts
- `csv_insight` - Analyze CSV data
- `improve_writing` - Improve text
- `review_code` - Code review

Output: Content + metadata

### Transform Modules (Tier 1 & 2)
- `extract_key_points` - Summarize content
- `translate` - Translate to language
- `format_markdown` - Format as markdown

Input: JSON from stdin
Output: Transformed JSON

### Sink Modules (Tier 1)
- `push_to_rag` - Add to RAG database
- `save_to_vault` - Save to Obsidian
- `export_pdf` - Convert to PDF

Input: JSON from stdin
Output: Success/failure status

## Implementation Pattern

```bash
#!/usr/bin/env bash
# module_name.sh

# ... standard setup ...

module.main() {
    local input="$1"

    # Process input
    local result=$(process_data "$input")

    # Detect pipe mode
    if [[ -t 1 ]]; then
        # Not piped - show human output
        echo "✓ Processed: $result"
        return 0
    else
        # Piped - output JSON
        jq -n \
            --arg status "success" \
            --arg module "module_name" \
            --arg result "$result" \
            '{
                status: $status,
                module: $module,
                data: {result: $result}
            }'
    fi
}
```

## Error Handling in Pipes

Errors always go to stderr, never break the pipe:

```bash
if [[ -z "$transcript" ]]; then
    util.log.error "Failed to extract transcript" >&2

    if [[ ! -t 1 ]]; then
        # Piped mode - output error JSON
        jq -n \
            --arg status "error" \
            --arg error "Failed to extract transcript" \
            '{status: $status, error: $error}'
    fi

    return 1
fi
```

## Benefits

1. **Composability** - Chain modules together
2. **Automation** - Script complex workflows
3. **Flexibility** - Mix and match processing steps
4. **Unix Philosophy** - Do one thing well, compose freely
5. **Backward Compatible** - Works standalone too

## RAG Database Schema

For `push_to_rag` module:

```sql
CREATE TABLE documents (
    id INTEGER PRIMARY KEY,
    source_module TEXT,
    source_type TEXT,  -- 'youtube', 'csv', 'code', etc.
    title TEXT,
    content TEXT,
    metadata JSON,
    embedding BLOB,
    created_at TIMESTAMP
);

CREATE TABLE chunks (
    id INTEGER PRIMARY KEY,
    document_id INTEGER,
    chunk_text TEXT,
    chunk_index INTEGER,
    embedding BLOB,
    FOREIGN KEY(document_id) REFERENCES documents(id)
);
```

## Next Steps

1. Update `youtube_to_obsidian` to output JSON when piped
2. Create `push_to_rag` module (Tier 1 - local SQLite)
3. Create `extract_key_points` transform module
4. Test piping workflows
