# BATTLESTAG Configuration

This directory contains configuration templates for the BATTLESTAG module system.

## Configuration Files

### llm.conf.example

Template for LLM (Language Model) configuration. This file controls both Tier 1 (local) and Tier 2 (cloud) LLM providers.

**Setup Instructions:**

1. **Copy the template to your home directory:**
   ```bash
   mkdir -p ~/.bstag/config
   cp llm.conf.example ~/.bstag/config/llm.conf
   chmod 600 ~/.bstag/config/llm.conf  # Protect API keys
   ```

2. **Edit ~/.bstag/config/llm.conf:**
   ```bash
   nano ~/.bstag/config/llm.conf
   # or
   vim ~/.bstag/config/llm.conf
   ```

3. **Configure Tier 1 (Local LLM):**
   - Set `LLM_ENDPOINT` to your Ollama/LM Studio endpoint
   - Set `LLM_MODEL` to your model name (e.g., `hermes3:8b`)
   - Adjust `LLM_MAX_TOKENS` and `LLM_TEMPERATURE` if needed

4. **Configure Tier 2 (Cloud LLM) - Optional:**
   - Choose a provider: `anthropic`, `openai`, `google`, `azure`, or `openrouter`
   - Set `TIER2_PROVIDER` to your chosen provider
   - Get an API key from your provider's website
   - Add your API key to the appropriate `*_API_KEY` variable

5. **Test your configuration:**
   ```bash
   # View current config
   bstag-module tier-info

   # Test Tier 1 (local)
   echo "Test text with some grammar error's." > /tmp/test.txt
   ~/.bstag/modules-available/tier1/text/improve_writing.sh /tmp/test.txt

   # Test Tier 2 (cloud) - if configured
   ~/.bstag/modules-available/tier2/wisdom/extract_wisdom.sh /tmp/article.txt
   ```

## Provider Selection Guide

### Anthropic (Claude) - Recommended
- **Best for:** High-quality reasoning, analysis, code review
- **Pricing:** ~$3/million input tokens, ~$15/million output tokens
- **Privacy:** Data sent to Anthropic (US-based)
- **Get API key:** https://console.anthropic.com/settings/keys

### OpenAI (GPT)
- **Best for:** General tasks, broad knowledge, creative writing
- **Pricing:** ~$2.50/million input tokens, ~$10/million output tokens
- **Privacy:** Data sent to OpenAI (US-based)
- **Get API key:** https://platform.openai.com/api-keys

### Google (Gemini)
- **Best for:** Free tier, multimodal tasks, cost-conscious users
- **Pricing:** Free tier available, then pay-as-you-go
- **Privacy:** Data sent to Google (US-based)
- **Get API key:** https://ai.google.dev/

### Azure OpenAI
- **Best for:** Enterprise deployments, compliance requirements
- **Pricing:** Similar to OpenAI, billed through Azure
- **Privacy:** Can be deployed in your Azure region
- **Get credentials:** https://portal.azure.com/

### OpenRouter
- **Best for:** Access to multiple models through one API
- **Pricing:** Varies by model ($3-15/million tokens typically)
- **Privacy:** Data sent to OpenRouter, then to provider
- **Get API key:** https://openrouter.ai/keys

## Privacy & Compliance

### Tier 1 (Local LLM)
- ✅ 100% private - data never leaves your infrastructure
- ✅ HIPAA/GDPR/SOC2 compliant (no data sharing)
- ✅ No API costs
- ✅ Works offline
- ⚠️ Limited capabilities compared to cloud models

### Tier 2 (Cloud LLM)
- ⚠️ Data sent to third-party provider
- ⚠️ Subject to provider's privacy policy
- ⚠️ May not be suitable for sensitive/regulated data
- ✅ Superior reasoning and analysis
- ✅ Better understanding of complex prompts
- 💰 Per-token pricing (~$0.01-0.10 per module execution)

## Cost Examples

Based on typical module usage:

| Module | Tier | Approx. Cost |
|--------|------|--------------|
| improve_writing | 1 | $0.00 |
| improve_writing | 2 | $0.01-0.02 |
| csv_insight | 1 | $0.00 |
| extract_wisdom | 2 | $0.03-0.05 |
| code_review (future) | 1 | $0.00 |
| code_review (future) | 2 | $0.05-0.10 |

**Monthly estimates (assuming 100 module executions):**
- Pure Tier 1: $0.00
- Mixed (80% Tier 1, 20% Tier 2): $2-5/month
- Pure Tier 2: $10-20/month

## Environment Variables

You can also set configuration via environment variables (useful for CI/CD):

```bash
# Tier 1
export LLM_ENDPOINT="http://192.168.1.232:11434/v1"
export LLM_MODEL="hermes3:8b"

# Tier 2
export TIER2_PROVIDER="anthropic"
export ANTHROPIC_API_KEY="sk-ant-..."

# Run module
bstag-module execute improve_writing /tmp/test.txt
```

Environment variables take precedence over config file settings.

## Security Best Practices

1. **Protect your config file:**
   ```bash
   chmod 600 ~/.bstag/config/llm.conf
   ```

2. **Never commit API keys to git:**
   - Config file is in `~/.bstag/` (not in repo)
   - `.gitignore` excludes all `*.conf` files

3. **Use separate API keys for different environments:**
   - Development key for testing
   - Production key for real work
   - Rotate keys periodically

4. **Monitor API usage:**
   - Most providers have usage dashboards
   - Set spending limits if available
   - Review monthly invoices

5. **Choose providers based on data sensitivity:**
   - Use Tier 1 for sensitive/regulated data
   - Use Tier 2 for non-sensitive analysis
   - Consider Azure for enterprise compliance needs

## Troubleshooting

### "Error: API key not set"
- Make sure you've created `~/.bstag/config/llm.conf`
- Verify the `*_API_KEY` variable is uncommented and has a value
- Check file permissions: `ls -la ~/.bstag/config/llm.conf`

### "Error: Unknown Tier 2 provider"
- Valid providers: `anthropic`, `openai`, `google`, `azure`, `openrouter`
- Check spelling in `TIER2_PROVIDER` variable

### "Error: No response" or "Error: ..."
- Check network connectivity
- Verify API key is valid (test on provider's website)
- Check endpoint URL is correct
- For Tier 1: Ensure Ollama/LM Studio is running

### Rate limits
- Most providers have rate limits (requests per minute)
- If you hit limits, add delays between module executions
- Consider upgrading to paid tier for higher limits

## Support

For issues or questions:
- GitHub Issues: https://github.com/battlestag/battlestag-bot/issues
- Documentation: See DOCS/ directory in repository
