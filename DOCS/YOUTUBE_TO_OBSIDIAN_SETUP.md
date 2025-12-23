# YouTube to Obsidian Setup Guide

Transform YouTube videos into comprehensive Obsidian notes with AI-powered wisdom extraction!

## What It Does

This Tier 2 module:
1. Downloads YouTube video transcripts
2. Extracts metadata (title, channel, duration, upload date)
3. Uses Claude to analyze and extract:
   - Summary
   - Key concepts
   - Actionable takeaways
   - Important quotes
   - Technical details
   - Resources mentioned
   - Questions to explore
4. Creates beautifully formatted Obsidian markdown notes
5. Includes full transcript in collapsible section

## Installation

### 1. Install yt-dlp

**Arch Linux:**
```bash
sudo pacman -S yt-dlp
```

**Debian/Ubuntu:**
```bash
sudo apt install yt-dlp
```

**macOS:**
```bash
brew install yt-dlp
```

**Pip (any OS):**
```bash
pip install --user yt-dlp
```

### 2. Configure Obsidian Vault Path

Edit your config file:
```bash
nano ~/.bstag/config/llm.conf
```

Add/update this line:
```bash
OBSIDIAN_VAULT_PATH="/path/to/your/obsidian/vault"
```

For example:
```bash
OBSIDIAN_VAULT_PATH="$HOME/Documents/MyVault"
OBSIDIAN_VAULT_PATH="$HOME/Obsidian/PersonalKnowledge"
OBSIDIAN_VAULT_PATH="/mnt/data/notes/ObsidianVault"
```

### 3. Install the Module

```bash
cd /path/to/BATTLESTAG-BOT/SCRIPTS
./install.sh --update
```

## Usage

### Basic Usage (with configured vault):
```bash
youtube_to_obsidian.sh "https://www.youtube.com/watch?v=VIDEO_ID"
```

### Specify vault path:
```bash
youtube_to_obsidian.sh "https://www.youtube.com/watch?v=VIDEO_ID" ~/Documents/MyVault
```

### Works with all YouTube URL formats:
```bash
youtube_to_obsidian.sh "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
youtube_to_obsidian.sh "https://youtu.be/dQw4w9WgXcQ"
youtube_to_obsidian.sh "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=120s"
```

## Example Output

The module creates a note like this:

```markdown
---
type: youtube-video
title: "How to Build a Second Brain"
channel: Ali Abdaal
url: https://www.youtube.com/watch?v=...
duration: 15:42
uploaded: 2024-01-15
processed: 2025-12-22 20:30:15
tags:
  - youtube
  - video-notes
---

# How to Build a Second Brain

**Channel**: [[Ali Abdaal]]
**URL**: https://www.youtube.com/watch?v=...
**Duration**: 15:42
**Uploaded**: 2024-01-15

---

## Summary
This video explains the concept of a "second brain" - an external system for capturing and organizing information...

## Key Concepts
- Progressive Summarization: highlight key points in stages
- CODE framework: Capture, Organize, Distill, Express
- Just-in-time organization vs just-in-case
...

## Actionable Takeaways
1. Start capturing ideas immediately when they occur
2. Use tags and folders sparingly - rely on search
3. Review notes weekly to reinforce learning
...

## Important Quotes
- "Your brain is for having ideas, not storing them"
- "The best note-taking system is the one you'll actually use"
...

[etc...]

---

<system-reminder>
Background Bash cd3fbd (command: bash -x modules/tier1/code/review_code.sh /tmp/test_code.py 2>&1 | tail -20) (status: running) Has new output available. You can check its output using the BashOutput tool.
</system-reminder>

## Transcript
<details>
<summary>Click to expand full transcript</summary>

[Full video transcript here...]

</details>
```

## Features

### Obsidian-Friendly
- YAML frontmatter with metadata
- Wikilinks for channel names (create backlinks)
- Tags for organization
- Collapsible transcript (doesn't clutter notes)
- Clean markdown formatting

### Smart Extraction
- Claude analyzes the entire transcript
- Pulls out actionable items (not just summary)
- Identifies tools, techniques, commands
- Captures memorable quotes
- Suggests follow-up questions

### Cost & Privacy
- **Cost**: ~$0.05-0.10 per video (based on transcript length)
- **Privacy**: Transcript and video metadata sent to Claude API
- **Speed**: 60-90 seconds for typical 10-20 minute video

## Tips & Tricks

### 1. Create Aliases
Add to your `~/.bashrc`:
```bash
alias yt2obs='youtube_to_obsidian.sh'
alias yt='youtube_to_obsidian.sh'
```

Then:
```bash
yt "https://youtu.be/VIDEO_ID"
```

### 2. Use with URL from Clipboard
```bash
# Linux (X11)
yt "$(xclip -o)"

# Linux (Wayland)
yt "$(wl-paste)"

# macOS
yt "$(pbpaste)"
```

### 3. Batch Process Playlist
```bash
# Get all video URLs from a playlist first, then:
while read -r url; do
    youtube_to_obsidian.sh "$url"
    sleep 5  # Rate limiting
done < playlist_urls.txt
```

### 4. Create Template Folder
In Obsidian, create a folder structure:
```
vault/
├── YouTube/
│   ├── Tech/
│   ├── Business/
│   ├── Health/
│   └── Learning/
```

Then specify the subfolder:
```bash
yt "URL" ~/vault/YouTube/Tech
```

### 5. Review Notes Weekly
Create an Obsidian Dataview query:
```dataview
TABLE
  channel as Channel,
  duration as Length,
  uploaded as "Upload Date"
FROM "YouTube"
WHERE type = "youtube-video"
SORT uploaded DESC
LIMIT 10
```

## Troubleshooting

### "yt-dlp not found"
- Install yt-dlp (see Installation section)
- Verify: `which yt-dlp`

### "Could not download transcript"
- Video may not have captions
- Try: `-dlp --list-subs "VIDEO_URL"` to check available subtitles
- Some videos have auto-generated captions only

### "Vault path does not exist"
- Check your `OBSIDIAN_VAULT_PATH` in config
- Verify the path exists: `ls ~/.bstag/config/llm.conf`

### "Error: Failed to extract wisdom"
- Claude API might be rate limited
- Check your API key is configured
- Try a shorter video first

## Advanced: Custom Prompts

Want different extraction format? Edit the prompt in:
```
~/.bstag/modules-available/tier2/obsidian/youtube_to_obsidian.sh
```

Look for the `prompt=` section around line 177.

## Examples

### Tech Tutorial:
```bash
yt "https://www.youtube.com/watch?v=..."
# Creates: "Complete Docker Tutorial for Beginners.md"
```

### Business/Self-Help:
```bash
yt "https://youtu.be/..."
# Creates: "7 Habits of Highly Effective People Summary.md"
```

### Course Lecture:
```bash
yt "https://www.youtube.com/watch?v=..."
# Creates: "MIT Algorithms Lecture 1 Introduction.md"
```

## What's Next?

Once you have notes in Obsidian:
1. Link related concepts with [[wikilinks]]
2. Add your own thoughts in "My Notes" section
3. Create MOCs (Maps of Content) linking related videos
4. Use Dataview to create dynamic indexes
5. Export to Anki flashcards for spaced repetition

---

**Pro Tip**: Use this for conference talks, tutorials, course lectures, podcasts (with video), interviews - anything with captions!

**Cost Estimate**: $5-10/month processes 50-100 videos (depending on length)
