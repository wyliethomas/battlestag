#!/usr/bin/env bash
# youtube_to_obsidian.sh - YouTube to Obsidian Notes (Tier 2)
#
# MODULE: youtube_to_obsidian
# NAME: YouTube to Obsidian
# CATEGORY: obsidian
# TIER: 2
# DESCRIPTION: Extracts YouTube transcript via HTTP, analyzes with Claude, saves to Obsidian vault
# DEPENDS: yt-dlp
# DEPENDS_PKG: arch:yt-dlp debian:yt-dlp fedora:yt-dlp brew:yt-dlp pip:yt-dlp
# CONFIG_REQUIRED: OBSIDIAN_VAULT_PATH
# CONFIG_PROMPT: "Enter the full path to your Obsidian vault"
# CONFIG_DEFAULT: ~/Documents/ObsidianVault
# PARAM: youtube_url:url:YouTube video URL
# PARAM: vault_path:path:Path to Obsidian vault (optional, uses config default)

set -euo pipefail

# Source utilities from .bstag installation
BSTAG_HOME="${BSTAG_HOME:-$HOME/.bstag}"

source "$BSTAG_HOME/utilities/file.sh"
source "$BSTAG_HOME/utilities/log.sh"
source "$BSTAG_HOME/utilities/llm.sh"

# Extract video ID from YouTube URL
extract_video_id() {
    local url="$1"

    # Handle various YouTube URL formats
    if [[ "$url" =~ youtu\.be/([a-zA-Z0-9_-]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    elif [[ "$url" =~ youtube\.com/watch\?v=([a-zA-Z0-9_-]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    elif [[ "$url" =~ youtube\.com/.*v=([a-zA-Z0-9_-]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    else
        echo ""
    fi
}

# Get video page HTML
get_video_page() {
    local video_id="$1"
    local url="https://www.youtube.com/watch?v=$video_id"

    curl -s -L "$url"
}

# Extract video title from HTML
extract_title() {
    local html="$1"

    # Try to extract from meta tags
    echo "$html" | grep -o '<meta name="title" content="[^"]*"' | head -1 | sed 's/.*content="\([^"]*\)".*/\1/' | sed 's/&quot;/"/g; s/&#39;/'"'"'/g; s/&amp;/\&/g'
}

# Extract channel name from HTML
extract_channel() {
    local html="$1"

    # Try to extract channel name
    echo "$html" | grep -o '"ownerChannelName":"[^"]*"' | head -1 | sed 's/"ownerChannelName":"\([^"]*\)"/\1/' | sed 's/\\u0026/\&/g'
}

# Extract video duration from HTML
extract_duration() {
    local html="$1"

    # Extract duration in seconds from schema
    echo "$html" | grep -o '"duration":"PT[^"]*"' | head -1 | sed 's/"duration":"PT\([^"]*\)"/\1/' | sed 's/H/*3600+/g; s/M/*60+/g; s/S//g' | bc 2>/dev/null || echo "0"
}

# Get transcript using yt-dlp
get_transcript() {
    local video_id="$1"
    local temp_dir=$(mktemp -d)

    util.log.info "Fetching transcript with yt-dlp..."

    # Use yt-dlp to get automatic captions
    yt-dlp --skip-download --write-auto-subs --sub-lang en --sub-format vtt \
        --output "$temp_dir/transcript.%(ext)s" \
        "https://www.youtube.com/watch?v=$video_id" >/dev/null 2>&1

    local transcript_file="$temp_dir/transcript.en.vtt"

    # If auto-captions fail, try manual subtitles
    if [[ ! -f "$transcript_file" ]]; then
        util.log.info "Auto-captions not available, trying manual subtitles..."
        yt-dlp --skip-download --write-subs --sub-lang en --sub-format vtt \
            --output "$temp_dir/transcript.%(ext)s" \
            "https://www.youtube.com/watch?v=$video_id" >/dev/null 2>&1
    fi

    if [[ ! -f "$transcript_file" ]]; then
        rm -rf "$temp_dir"
        echo "Error: Could not extract transcript (no captions available)" >&2
        return 1
    fi

    # Parse VTT format - extract only the text, remove timestamps and tags
    local transcript
    transcript=$(grep -v '^WEBVTT' "$transcript_file" | \
        grep -v '^Kind:' | \
        grep -v '^Language:' | \
        grep -v '^$' | \
        grep -v '^[0-9][0-9]:' | \
        sed 's/<[^>]*>//g' | \
        tr '\n' ' ' | \
        sed 's/  */ /g')

    rm -rf "$temp_dir"
    echo "$transcript"
}

# Format seconds to HH:MM:SS or MM:SS
format_duration() {
    local seconds="$1"

    if [[ -z "$seconds" ]] || [[ "$seconds" == "0" ]]; then
        echo "unknown"
        return
    fi

    local hours=$((seconds / 3600))
    local minutes=$(( (seconds % 3600) / 60 ))
    local secs=$((seconds % 60))

    if [[ $hours -gt 0 ]]; then
        printf "%d:%02d:%02d" $hours $minutes $secs
    else
        printf "%d:%02d" $minutes $secs
    fi
}

# Module: youtube_to_obsidian
# Convert YouTube video to Obsidian note with Claude wisdom extraction
module.youtube_to_obsidian() {
    local youtube_url="${1:-}"
    local vault_path="${2:-}"

    # Validation
    if [[ -z "$youtube_url" ]]; then
        echo "Error: YouTube URL required" >&2
        echo "Usage: youtube_to_obsidian.sh <youtube_url> [vault_path]" >&2
        return 1
    fi

    # Get default vault path from config if not provided
    if [[ -z "$vault_path" ]]; then
        if [[ -f "$BSTAG_HOME/config/llm.conf" ]]; then
            source "$BSTAG_HOME/config/llm.conf"
            vault_path="${OBSIDIAN_VAULT_PATH:-}"
        fi

        if [[ -z "$vault_path" ]]; then
            echo "Error: Vault path not specified and no default configured" >&2
            echo "Either:" >&2
            echo "  1. Pass vault path: youtube_to_obsidian.sh <url> /path/to/vault" >&2
            echo "  2. Set OBSIDIAN_VAULT_PATH in ~/.bstag/config/llm.conf" >&2
            return 1
        fi
    fi

    # Validate vault path exists
    if [[ ! -d "$vault_path" ]]; then
        echo "Error: Vault path does not exist: $vault_path" >&2
        return 1
    fi

    util.log.info "Processing YouTube video..."
    util.log.info "Vault: $vault_path"

    # Extract video ID
    local video_id
    video_id=$(extract_video_id "$youtube_url")

    if [[ -z "$video_id" ]]; then
        echo "Error: Could not extract video ID from URL: $youtube_url" >&2
        return 1
    fi

    util.log.info "Video ID: $video_id"

    # Get video page for metadata
    util.log.info "Fetching metadata..."
    local html
    html=$(get_video_page "$video_id")

    # Extract metadata
    local title
    title=$(extract_title "$html")
    local channel
    channel=$(extract_channel "$html")
    local duration_seconds
    duration_seconds=$(extract_duration "$html")
    local duration_formatted
    duration_formatted=$(format_duration "$duration_seconds")

    if [[ -z "$title" ]]; then
        title="YouTube Video $video_id"
    fi

    util.log.info "Title: $title"
    util.log.info "Channel: $channel"
    util.log.info "Duration: $duration_formatted"

    # Get transcript
    local transcript
    transcript=$(get_transcript "$video_id")

    if [[ -z "$transcript" ]]; then
        echo "Error: Could not extract transcript" >&2
        return 1
    fi

    local transcript_length
    transcript_length=$(echo "$transcript" | wc -w)

    util.log.info "Transcript: $transcript_length words"

    # Extract wisdom with Claude (Tier 2)
    util.log.info "Extracting wisdom with Claude (60-90 seconds)..."

    local prompt="Analyze this YouTube video transcript and create comprehensive notes.

Video: $title
Channel: $channel
Duration: $duration_formatted

Extract:

## Summary
[2-3 sentence overview of the video's main message]

## Key Concepts
[5-10 main concepts or ideas, as bullet points]

## Actionable Takeaways
[Specific actions the viewer should take, numbered list]

## Important Quotes
[5-10 memorable or impactful quotes from the video]

## Technical Details
[Any specific tools, techniques, commands, or technical information mentioned]

## Resources Mentioned
[Books, websites, tools, or other resources referenced]

## My Notes
[Leave this section empty with placeholder text: \"Add your personal notes and thoughts here...\"]

## Questions to Explore
[3-5 questions this video raises or topics to research further]

Format as clean, well-organized markdown suitable for Obsidian.

Transcript:
---
$transcript
---"

    local wisdom
    wisdom=$(util.llm.query_big "$prompt" 4096)

    # Check if we got a response
    if [[ -z "$wisdom" ]] || [[ "$wisdom" == "Error:"* ]]; then
        echo "Error: Failed to extract wisdom" >&2
        echo "Response: $wisdom" >&2
        return 1
    fi

    # Create safe filename
    local safe_title
    safe_title=$(echo "$title" | sed 's/[^a-zA-Z0-9 ]//g' | sed 's/  */ /g' | cut -c1-80)
    local note_filename="${safe_title}.md"
    local note_path="$vault_path/$note_filename"

    # Create Obsidian note
    local timestamp
    timestamp=$(date -u +"%Y-%m-%d %H:%M:%S")
    local upload_date
    upload_date=$(date +"%Y-%m-%d")

    cat > "$note_path" << EOF
---
type: youtube-video
title: "$title"
channel: $channel
url: $youtube_url
video_id: $video_id
duration: $duration_formatted
processed: $timestamp
tags:
  - youtube
  - video-notes
---

# $title

**Channel**: [[$channel]]
**URL**: $youtube_url
**Duration**: $duration_formatted

---

$wisdom

---

## Transcript
<details>
<summary>Click to expand full transcript ($transcript_length words)</summary>

$transcript

</details>

---

*Processed with BATTLESTAG YouTube to Obsidian (Tier 2)*
*Cost: ~\$0.05-0.10 | No external dependencies required*
EOF

    util.log.info "✓ Note saved to Obsidian vault"

    # Read the created file content for piping
    local file_content
    file_content=$(cat "$note_path")

    # Check if output is piped
    if [[ -t 1 ]]; then
        # Not piped - show human-readable output to terminal
        echo ""
        echo "Summary:"
        echo "  Video:     $title"
        echo "  Channel:   $channel"
        echo "  Duration:  $duration_formatted"
        echo "  Words:     $transcript_length"
        echo "  Vault:     $vault_path"
        echo "  File:      $note_filename"
        echo "  Tier:      2 (Claude Analysis)"
        echo "  Cost:      ~\$0.05-0.10"
        echo ""
        echo "✓ Done! Open in Obsidian: $note_path"
    else
        # Piped - output JSON to stdout for next command in pipeline
        jq -n \
            --arg status "success" \
            --arg module "youtube_to_obsidian" \
            --argjson tier 2 \
            --arg file_path "$note_path" \
            --arg title "$title" \
            --arg channel "$channel" \
            --arg duration "$duration_formatted" \
            --arg video_id "$video_id" \
            --arg url "$youtube_url" \
            --argjson word_count "$transcript_length" \
            --arg content "$file_content" \
            --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
            '{
                status: $status,
                module: $module,
                tier: $tier,
                data: {
                    file_path: $file_path,
                    title: $title,
                    content: $content,
                    metadata: {
                        video_id: $video_id,
                        channel: $channel,
                        duration: $duration,
                        url: $url,
                        word_count: $word_count
                    }
                },
                cost: 0.08,
                timestamp: $timestamp
            }'
    fi
}

# Execute if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    module.youtube_to_obsidian "$@"
fi
