#!/usr/bin/env bash
# push_to_rag.sh - Push content to RAG database (Tier 1)
#
# MODULE: push_to_rag
# NAME: Push to RAG Database
# CATEGORY: rag
# TIER: 1
# DESCRIPTION: Accepts piped JSON input and adds content to local RAG database for retrieval
# DEPENDS: sqlite3
# DEPENDS_PKG: arch:sqlite debian:sqlite3 fedora:sqlite brew:sqlite pip:sqlite3
# CONFIG_REQUIRED: RAG_DB_PATH
# CONFIG_PROMPT: "Enter the path for your RAG database"
# CONFIG_DEFAULT: ~/.local/share/battlestag/rag.db
# PARAM: input:json:JSON input from piped module (optional, reads from stdin)

set -euo pipefail

# Source utilities from .bstag installation
BSTAG_HOME="${BSTAG_HOME:-$HOME/.bstag}"

source "$BSTAG_HOME/utilities/file.sh"
source "$BSTAG_HOME/utilities/log.sh"

# Initialize RAG database
init_rag_db() {
    local db_path="$1"

    # Create directory if it doesn't exist
    mkdir -p "$(dirname "$db_path")"

    # Create tables if they don't exist
    sqlite3 "$db_path" << 'EOF'
CREATE TABLE IF NOT EXISTS documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_module TEXT NOT NULL,
    source_type TEXT NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chunks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    document_id INTEGER NOT NULL,
    chunk_text TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(document_id) REFERENCES documents(id)
);

CREATE INDEX IF NOT EXISTS idx_documents_source ON documents(source_module, source_type);
CREATE INDEX IF NOT EXISTS idx_documents_created ON documents(created_at);
CREATE INDEX IF NOT EXISTS idx_chunks_document ON chunks(document_id);
EOF

    util.log.info "RAG database initialized: $db_path"
}

# Chunk text into smaller pieces for better retrieval
chunk_text() {
    local text="$1"
    local chunk_size="${2:-1000}"  # Default 1000 characters per chunk

    # Simple chunking by character count (can be improved with sentence boundaries)
    echo "$text" | fold -w "$chunk_size" -s
}

# Add document to RAG database
add_to_rag() {
    local json_input="$1"
    local db_path="$2"

    # Extract fields from JSON
    local source_module
    source_module=$(echo "$json_input" | jq -r '.module // "unknown"')

    local title
    title=$(echo "$json_input" | jq -r '.data.title // "Untitled"')

    local content
    content=$(echo "$json_input" | jq -r '.data.content // empty')

    if [[ -z "$content" ]]; then
        util.log.error "No content found in JSON input"
        return 1
    fi

    local metadata
    metadata=$(echo "$json_input" | jq -c '.data.metadata // {}')

    # Determine source type from metadata or module name
    local source_type
    if echo "$metadata" | jq -e '.video_id' >/dev/null 2>&1; then
        source_type="youtube"
    elif echo "$metadata" | jq -e '.file_type' >/dev/null 2>&1; then
        source_type="file"
    else
        source_type="unknown"
    fi

    util.log.info "Adding document to RAG database..."
    util.log.info "  Module: $source_module"
    util.log.info "  Type: $source_type"
    util.log.info "  Title: $title"

    # Insert document
    local doc_id
    doc_id=$(sqlite3 "$db_path" << EOF
INSERT INTO documents (source_module, source_type, title, content, metadata)
VALUES (
    '$source_module',
    '$source_type',
    '$(echo "$title" | sed "s/'/''/g")',
    '$(echo "$content" | sed "s/'/''/g")',
    '$(echo "$metadata" | sed "s/'/''/g")'
);
SELECT last_insert_rowid();
EOF
)

    util.log.info "  Document ID: $doc_id"

    # Chunk the content
    local chunk_index=0
    while IFS= read -r chunk; do
        [[ -z "$chunk" ]] && continue

        sqlite3 "$db_path" << EOF
INSERT INTO chunks (document_id, chunk_text, chunk_index)
VALUES (
    $doc_id,
    '$(echo "$chunk" | sed "s/'/''/g")',
    $chunk_index
);
EOF

        chunk_index=$((chunk_index + 1))
    done < <(chunk_text "$content" 1000)

    util.log.info "  Chunks created: $chunk_index"

    # Return stats
    echo "$doc_id|$chunk_index"
}

# Module: push_to_rag
# Accept piped JSON and add to RAG database
module.push_to_rag() {
    local input="${1:-}"

    # Get RAG database path from config
    local db_path
    if [[ -f "$BSTAG_HOME/config/llm.conf" ]]; then
        source "$BSTAG_HOME/config/llm.conf"
        db_path="${RAG_DB_PATH:-}"
    fi

    if [[ -z "$db_path" ]]; then
        db_path="$HOME/.local/share/battlestag/rag.db"
        util.log.warn "RAG_DB_PATH not configured, using default: $db_path"
    fi

    # Initialize database if it doesn't exist
    if [[ ! -f "$db_path" ]]; then
        init_rag_db "$db_path"
    fi

    # Read JSON from stdin if not provided as argument
    local json_input
    if [[ -n "$input" ]]; then
        json_input="$input"
    else
        json_input=$(cat)
    fi

    # Validate JSON
    if ! echo "$json_input" | jq empty 2>/dev/null; then
        util.log.error "Invalid JSON input"
        return 1
    fi

    # Check for error status in input
    local status
    status=$(echo "$json_input" | jq -r '.status // "unknown"')

    if [[ "$status" == "error" ]]; then
        local error_msg
        error_msg=$(echo "$json_input" | jq -r '.error // "Unknown error"')
        util.log.error "Previous module failed: $error_msg"
        return 1
    fi

    # Add to RAG database
    local result
    result=$(add_to_rag "$json_input" "$db_path")

    local doc_id chunk_count
    IFS='|' read -r doc_id chunk_count <<< "$result"

    # Check if output is piped
    if [[ -t 1 ]]; then
        # Not piped - show human-readable output
        echo ""
        echo "✓ Added to RAG database"
        echo "  Database: $db_path"
        echo "  Document ID: $doc_id"
        echo "  Chunks: $chunk_count"
        echo ""
    else
        # Piped - output JSON for further processing
        jq -n \
            --arg status "success" \
            --arg module "push_to_rag" \
            --argjson tier 1 \
            --arg db_path "$db_path" \
            --argjson doc_id "$doc_id" \
            --argjson chunk_count "$chunk_count" \
            --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
            '{
                status: $status,
                module: $module,
                tier: $tier,
                data: {
                    db_path: $db_path,
                    document_id: $doc_id,
                    chunk_count: $chunk_count
                },
                timestamp: $timestamp
            }'
    fi
}

# Execute if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    module.push_to_rag "$@"
fi
