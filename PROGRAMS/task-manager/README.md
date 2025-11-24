# Task Manager

A hierarchical project management system designed for tracking long-running "epic" tasks across multiple life domains. Built for the Battlestag Agent ecosystem.

## Overview

This is NOT a todo list with reminders - it's a project journal with intelligent prioritization for managing nested, contextual work. Track projects hierarchically under contexts, maintain journal logs for each project, and manage granular checklist items.

## Key Concepts

- **Contexts**: Top-level categories (property, house, pxp, personal)
- **Projects**: Ongoing initiatives within contexts (e.g., "Sawmill" under "property")
- **Checklist Items**: Granular tasks within projects
- **Journal Entries**: Narrative log of what happened, blockers, and next steps
- **Intelligent Queries**: LLM-friendly JSON output for context-aware suggestions

## Installation

```bash
cd PROGRAMS/task-manager
./install.sh
```

The installer will:
- Build all four executables
- Install to `~/.local/bin`
- Create wrapper scripts with environment loading
- Seed default contexts (property, house, pxp, personal)
- Update your PATH if needed

## Executables

All executables use the `task_` prefix (with underscore):

- **task_project** - Manage projects and contexts
- **task_checklist** - Manage checklist items within projects
- **task_journal** - Add/view journal entries
- **task_query** - Intelligent queries with JSON output

Use the `*_run` wrapper scripts for automatic environment loading.

## Quick Start

```bash
# List contexts
task_project_run contexts --list

# Add a new project
task_project_run add --name "Sawmill" --context "property" --goal "Fully operational sawmill"

# Add checklist items (supports batch adding)
task_checklist_run add --project 1 --items "Level rails, Test cut, Build outfeed table"

# Log work
task_journal_run add --project 1 --entry "Got the blade mounted but rails are out of level"

# Check off an item
task_checklist_run check --id 1

# View project status (JSON for LLM)
task_query_run status --project 1

# Get suggestions for what to work on
task_query_run suggest --context "property" --time "3 days"
```

## Usage Examples

### Managing Projects

```bash
# Add a new context
task_project_run contexts --add "workshop"

# List all projects
task_project_run list

# Filter by context
task_project_run list --context "property"

# Filter by status
task_project_run list --status active

# Update project status
task_project_run update --id 1 --status blocked

# Update project goal
task_project_run update --id 1 --goal "Updated goal description"

# Archive completed project
task_project_run archive --id 1
```

### Managing Checklists

```bash
# Add single item
task_checklist_run add --project 1 --item "Level the carriage rails"

# Add multiple items at once
task_checklist_run add --project 1 --items "Order blade, Research shimming, Pick up lumber"

# List all items for a project
task_checklist_run list --project 1

# List only incomplete items
task_checklist_run list --project 1 --pending

# Mark item complete
task_checklist_run check --id 5

# Mark item incomplete
task_checklist_run uncheck --id 5

# Remove an item
task_checklist_run remove --id 5

# Reorder an item
task_checklist_run reorder --id 5 --position 1
```

### Journal Entries

```bash
# Add an entry
task_journal_run add --project 1 --entry "Spent 2 hours today. Rails need shimming."

# List all entries
task_journal_run list --project 1

# Limit to recent entries
task_journal_run list --project 1 --limit 5

# View latest entry
task_journal_run latest --project 1
```

### Intelligent Queries

All query commands output JSON for easy LLM consumption:

```bash
# Get full project status
task_query_run status --project 1
task_query_run status --name "sawmill"

# Overview of all active projects
task_query_run overview
task_query_run overview --context "property"

# Get suggestions (considers completion momentum and staleness)
task_query_run suggest --context "property" --time "weekend"

# Find neglected projects
task_query_run neglected --days 14

# Find projects near completion
task_query_run closing --threshold 70

# Context view with stats
task_query_run context --name "property"
```

## Database Schema

Located at `~/.local/share/task-manager/tasks.db`

### Tables

- **contexts**: Top-level categories
- **projects**: Projects within contexts
- **checklist_items**: Granular tasks within projects
- **journal_entries**: Narrative logs for projects

### Key Features

- Progress calculation: `(done_items / total_items) * 100`
- `last_touched` auto-updates when:
  - Journal entry is added
  - Checklist item is added, checked, or unchecked
  - Project status/goal is updated

## Configuration

- Database: `~/.local/share/task-manager/tasks.db`
- Config: `~/.config/task-manager/.env`
- Override database path: `export DB_PATH=/custom/path/tasks.db`

## Exit Codes

- `0` - Success
- `1` - Invalid arguments
- `2` - Database error
- `3` - Not found (project, context, item, entry)

## Architecture

Built following the financial-asset-tracker pattern:

```
task-manager/
├── cmd/                    # Executable entry points
│   ├── project/
│   ├── checklist/
│   ├── journal/
│   └── query/
├── config/                 # Configuration loading
├── db/                     # Database layer
├── pkg/
│   ├── app/               # Initialization helpers
│   └── exitcodes/         # Standard exit codes
├── install.sh
├── uninstall.sh
└── update.sh
```

## Update

```bash
cd PROGRAMS/task-manager
./update.sh
```

Preserves configuration and data while updating binaries.

## Uninstall

```bash
cd PROGRAMS/task-manager
./uninstall.sh
```

Removes binaries and configuration. Optionally removes database.

## Integration with Battlestag Agent

The `task_query` executable outputs JSON specifically formatted for LLM consumption. The agent can:

1. Query project status for context
2. Get suggestions based on available time and context
3. Identify neglected or nearly-complete projects
4. Access full project history via journal entries

Example agent interaction:
```
User: "I'm at the property for 3 days, what should I focus on?"
Agent: [Runs task_query suggest --context "property" --time "3 days"]
Agent: Based on your projects, I suggest focusing on the Sawmill (65% complete)...
```

## Development

### Building

```bash
go build -o task_project ./cmd/project
go build -o task_checklist ./cmd/checklist
go build -o task_journal ./cmd/journal
go build -o task_query ./cmd/query
```

### Testing

```bash
# Set test database
export DB_PATH=/tmp/test-tasks.db

# Run test workflow
task_project contexts --add "test"
task_project add --name "Test Project" --context "test" --goal "Test goal"
task_checklist add --project 1 --items "Item 1, Item 2, Item 3"
task_journal add --project 1 --entry "Test entry"
task_checklist check --id 1
task_query status --project 1
```

## License

Part of the BATTLESTAG-BOT project.
