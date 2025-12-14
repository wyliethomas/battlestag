# Setting Up Claude API with OpenCode

## Quick Reference for After Lunch

### Step 1: Get Claude API Key

1. Go to: https://console.anthropic.com/
2. Sign up or log in
3. Navigate to API Keys section
4. Create a new API key
5. Copy the key (starts with `sk-ant-...`)

**Pricing:** Claude Sonnet 3.5 is ~$3 per million input tokens, $15 per million output tokens
- Very affordable for personal use
- Much cheaper than GPT-4

### Step 2: Configure OpenCode on Battlestag

```bash
ssh battlestag

# Add Claude authentication
~/.opencode/bin/opencode auth
# Select "Claude" and paste your API key

# Update opencode.json to include Claude provider
```

### Step 3: Update Configuration

Edit `~/.config/opencode/opencode.json` on battlestag:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "anthropic": {
      "npm": "@ai-sdk/anthropic",
      "name": "Anthropic",
      "models": {
        "claude-sonnet-4": {
          "name": "Claude Sonnet 4"
        },
        "claude-sonnet-3-5": {
          "name": "Claude Sonnet 3.5"
        }
      }
    },
    "ollama": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Ollama (battlestag)",
      "options": {
        "baseURL": "http://192.168.1.232:11434/v1"
      },
      "models": {
        "llama3.1:8b-32k": {
          "name": "Llama 3.1 8B (32k context)"
        },
        "qwen2.5-coder:7b": {
          "name": "Qwen 2.5 Coder 7B"
        }
      }
    }
  }
}
```

### Step 4: Test with Rails Automation

```bash
cd ~/projects/rails-automation-test

# Option A: CLI mode (should work now!)
~/.opencode/bin/opencode run --model anthropic/claude-sonnet-3-5 --agent=build "$(cat rails-prompt.txt)"

# Option B: TUI mode
~/.opencode/bin/opencode --model anthropic/claude-sonnet-3-5
# Then paste the Rails prompt
```

### Step 5: Verify It Works

```bash
# Check if files were created
ls -la ~/projects/rails-automation-test/

# Should see:
# - test-rails-automation/ directory
# - Rails template extracted and configured
# - .env file updated
# - setup.sh made executable
```

## Expected Behavior with Claude

With Claude API, you should see:
- ✅ Tool calls execute properly
- ✅ Files actually get created/modified
- ✅ Bash commands run successfully
- ✅ Multi-step automation works smoothly
- ✅ Much better reasoning and planning

## Commands Quick Reference

```bash
# List available models
~/.opencode/bin/opencode models

# Check authentication
~/.opencode/bin/opencode auth

# View sessions
~/.opencode/bin/opencode session list

# Export session for review
~/.opencode/bin/opencode export SESSION_ID
```

## Next Steps After Testing

Once Rails automation works with Claude:

1. Document the successful approach
2. Create reusable project templates
3. Build automation scripts that call OpenCode with Claude
4. Test other project types (Go, Python, Node.js, etc.)

---

**Status:** Ready to set up after lunch!
**Files Ready:** Rails prompt at `~/projects/rails-automation-test/rails-prompt.txt`
