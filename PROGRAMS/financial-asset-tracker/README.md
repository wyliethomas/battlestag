# Financial Asset Tracker

A lightweight, privacy-focused financial asset tracking system that runs locally on your machine. Track vehicles, property, investments, and other assets with full value history.

## Features

- **Multiple Asset Types**: Track vehicles, property, investments, and custom categories
- **Full Value History**: Every value update is recorded with timestamps
- **Soft Delete**: Removed assets are preserved with history for record-keeping
- **Stale Asset Detection**: Get warnings for assets that haven't been updated recently
- **Privacy-First**: All data stored locally in SQLite - no cloud, no tracking
- **Multiple Output Formats**: JSON and CSV export options
- **Lightweight**: Two small executables, no background processes

## Architecture

### Executables

1. **financial-asset-tracker** - Management tool for adding, updating, and removing assets
2. **financial-asset-tracker-query** - Read-only query tool for listing and analyzing assets

### Database

SQLite database storing:
- **Assets**: Core asset information with current values
- **Value History**: Complete timeline of all value changes

Default location: `~/.local/share/financial-asset-tracker/assets.db`

## Installation

```bash
./install.sh
```

The installer will prompt for:
- Installation directory (default: `~/.local/bin`)
- Configuration directory (default: `~/.config/financial-asset-tracker`)
- Database file path (default: `~/.local/share/financial-asset-tracker/assets.db`)

Make sure your installation directory is in your `PATH`.

## Usage

### Adding Assets

```bash
# Add a vehicle with full purchase info
financial-asset-tracker-run add \
  --name "2019 Honda Civic" \
  --category vehicle \
  --purchase-price 25000 \
  --purchase-date 2019-06-15 \
  --current-value 18000

# Add property without purchase info
financial-asset-tracker-run add \
  --name "Main Residence" \
  --category property \
  --current-value 450000

# Add investment
financial-asset-tracker-run add \
  --name "S&P 500 Index Fund" \
  --category investment \
  --purchase-price 10000 \
  --purchase-date 2023-01-15 \
  --current-value 12500 \
  --notes "Retirement account"
```

### Updating Asset Values

```bash
# Update asset value
financial-asset-tracker-run update --id 1 --value 17500

# Update with notes
financial-asset-tracker-run update --id 1 --value 17500 --notes "Post-accident valuation"
```

### Removing Assets

```bash
# Soft delete (preserves history)
financial-asset-tracker-run remove --id 1

# Remove with specific date
financial-asset-tracker-run remove --id 1 --date 2024-11-15
```

### Restoring Assets

```bash
# Restore a removed asset
financial-asset-tracker-run restore --id 1
```

### Querying Assets

```bash
# List all active assets
financial-asset-tracker-query-run --pretty

# Include removed assets
financial-asset-tracker-query-run --all --pretty

# Filter by category
financial-asset-tracker-query-run --category vehicle --pretty

# Check for stale assets (not updated in 30 days)
financial-asset-tracker-query-run --stale-days 30 --pretty

# Show value history for an asset
financial-asset-tracker-query-run --history --id 1 --pretty

# Show summary statistics
financial-asset-tracker-query-run --summary --pretty

# Export to CSV
financial-asset-tracker-query-run --csv > assets.csv
```

## Output Examples

### List Assets (JSON)

```json
{
  "total_count": 3,
  "assets": [
    {
      "id": 1,
      "name": "2019 Honda Civic",
      "category": "vehicle",
      "purchase_price": 25000,
      "purchase_date": "2019-06-15T00:00:00Z",
      "current_value": 18000,
      "date_added": "2024-11-18T10:00:00Z",
      "last_updated": "2024-11-18T10:00:00Z",
      "is_removed": false,
      "notes": ""
    }
  ],
  "warnings": [
    "Asset #2 'Investment Portfolio' not updated in 45 days"
  ]
}
```

### Value History

```json
{
  "asset": {
    "id": 1,
    "name": "2019 Honda Civic",
    "category": "vehicle",
    "current_value": 17500
  },
  "history": [
    {
      "id": 3,
      "asset_id": 1,
      "value": 17500,
      "recorded_date": "2024-11-15T00:00:00Z",
      "notes": "Post-accident valuation",
      "created_at": "2024-11-15T14:30:00Z"
    },
    {
      "id": 2,
      "asset_id": 1,
      "value": 18000,
      "recorded_date": "2024-10-01T00:00:00Z",
      "notes": "Annual update",
      "created_at": "2024-10-01T09:00:00Z"
    },
    {
      "id": 1,
      "asset_id": 1,
      "value": 25000,
      "recorded_date": "2019-06-15T00:00:00Z",
      "notes": "Initial value",
      "created_at": "2019-06-15T12:00:00Z"
    }
  ]
}
```

### Summary Statistics

```json
{
  "total_value": 480500,
  "total_count": 3,
  "categories": [
    {
      "category": "property",
      "count": 1,
      "value": 450000
    },
    {
      "category": "vehicle",
      "count": 1,
      "value": 18000
    },
    {
      "category": "investment",
      "count": 1,
      "value": 12500
    }
  ]
}
```

## Database Schema

### Assets Table

| Field | Type | Description |
|-------|------|-------------|
| id | INTEGER | Primary key |
| name | TEXT | Asset name |
| category | TEXT | vehicle, property, investment, other |
| purchase_price | REAL | Optional purchase price |
| purchase_date | DATE | Optional purchase date |
| current_value | REAL | Current value |
| date_added | DATETIME | When asset was added |
| last_updated | DATETIME | When value was last updated |
| is_removed | BOOLEAN | Soft delete flag |
| removed_date | DATE | When asset was removed |
| notes | TEXT | Additional notes |

### Asset Value History Table

| Field | Type | Description |
|-------|------|-------------|
| id | INTEGER | Primary key |
| asset_id | INTEGER | Foreign key to assets |
| value | REAL | Value at this point in time |
| recorded_date | DATE | Date this value was recorded |
| notes | TEXT | Notes about this update |
| created_at | DATETIME | When this record was created |

## Configuration

Environment variables (set in `~/.config/financial-asset-tracker/.env`):

```bash
# Database file location
DB_PATH=~/.local/share/financial-asset-tracker/assets.db
```

## Exit Codes

Both executables use standard exit codes:

- `0` - Success
- `1` - Invalid arguments
- `2` - Database error
- `3` - Asset not found

## Updates

To update after pulling new code:

```bash
git pull
./update.sh
```

## Uninstallation

```bash
./uninstall.sh
```

The uninstaller will prompt you to:
- Remove wrapper scripts
- Remove configuration directory
- Remove database (THIS DELETES ALL DATA)

## Use Cases

### Personal Finance Tracking
Track net worth by monitoring all assets in one place.

### Insurance Records
Maintain up-to-date valuations for insurance purposes.

### Investment Monitoring
Track investment performance over time with full history.

### Estate Planning
Keep accurate records of all assets for estate planning.

### Tax Preparation
Export asset data for tax reporting and capital gains calculations.

## Privacy & Security

- **100% Local**: All data stays on your machine
- **No Network Calls**: Executables never connect to the internet
- **SQLite Storage**: Industry-standard, auditable database format
- **Open Source**: Review the code yourself

## Integration

### With Other Financial Tools

```bash
# Query current assets and pipe to other tools
financial-asset-tracker-query-run | jq '.assets[] | select(.category == "investment")'

# Export to CSV for spreadsheet analysis
financial-asset-tracker-query-run --csv > assets.csv
```

### Automated Updates

Create a cron job or systemd timer to remind yourself to update values:

```bash
# Check for stale assets daily at 9 AM
0 9 * * * financial-asset-tracker-query-run --stale-days 30 --pretty
```

## Troubleshooting

### Database locked

If you see "database locked" errors:
- Make sure only one process is writing to the database
- The query tool can run concurrently (read-only)

### Permission denied

Make sure wrapper scripts are executable:
```bash
chmod +x ~/.local/bin/financial-asset-tracker*
```

### Command not found

Ensure your installation directory is in PATH:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## Contributing

This is a personal finance tool. If you find bugs or have feature requests, please open an issue.

## License

MIT License - See LICENSE file for details

## Future Enhancements

Potential features for future versions:
- Import from CSV
- Depreciation calculations
- Asset appreciation trends
- Category customization
- Multi-currency support
- Backup/restore utilities
