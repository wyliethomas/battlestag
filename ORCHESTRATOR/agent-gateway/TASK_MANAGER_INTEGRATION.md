# Task Manager API Integration

## Summary

The task-manager program has been successfully integrated into the agent-gateway API. The LLM can now discover and use all task management functionality automatically.

## Files Changed

### 1. New File: `programs/task_manager.go`
**Location**: `/home/battlestag/Work/BATTLESTAG-BOT/ORCHESTRATOR/agent-gateway/programs/task_manager.go`

This is a thin adapter that implements the `Program` interface. It:
- Provides metadata about the task-manager to the LLM
- Translates LLM requests into calls to your installed executables (`task_project_run`, `task_checklist_run`, `task_journal_run`, `task_query_run`)
- Handles parameter validation and command construction

**Key Features**:
- 17 different commands exposed to the LLM
- All 4 task manager executables accessible
- Automatic JSON formatting for query commands
- Comprehensive parameter validation

### 2. Modified: `main.go` (lines 102-106)
**Location**: `/home/battlestag/Work/BATTLESTAG-BOT/ORCHESTRATOR/agent-gateway/main.go`

Added registration of the task-manager program:
```go
if err := programsRegistry.Register(programs.NewTaskManagerProgram()); err != nil {
    log.Printf("Warning: Failed to register task-manager program: %v", err)
} else {
    log.Println("Registered program: task-manager")
}
```

## Available Commands

The LLM now has access to these task-manager commands:

### Context Management
- `list-contexts` - List all contexts
- `add-context` - Add a new context

### Project Management
- `add-project` - Create a new project
- `list-projects` - List projects (with optional filters)
- `update-project` - Update project status or goal

### Checklist Management
- `add-checklist` - Add checklist items (supports batch)
- `list-checklist` - List checklist items
- `check-item` - Mark item complete
- `uncheck-item` - Mark item incomplete

### Journal Management
- `add-journal` - Add journal entry
- `list-journal` - List journal entries
- `latest-journal` - Get most recent entry

### Intelligent Queries (JSON output)
- `query-status` - Full project status
- `query-overview` - Overview of all/filtered projects
- `query-suggest` - AI-friendly suggestions based on context/time
- `query-neglected` - Projects not touched recently
- `query-closing` - Projects near completion
- `query-context` - All projects in a context with stats

## How the LLM Uses It

When the agent-gateway starts, the `buildProgramsContext()` function automatically injects this information into the system prompt:

```
You have access to the following programs that can perform tasks:

- Task Manager (task-manager): Hierarchical project management system for tracking
  long-running tasks across life domains (property, house, pxp, personal).
  Manages projects with journal logs and checklists. Use this for task tracking,
  project status, and intelligent work suggestions.

  Parameters:
    • command (string) [REQUIRED]: Command to execute: 'list-contexts', 'add-context', ...
    • context (string): Context name (property, house, pxp, personal) ...
    • project_id (int): Project ID ...
    [... etc ...]
```

## Example LLM Interaction Flow

**User**: "Show me all my property projects"

**LLM**:
```
EXECUTE_PROGRAM: task-manager
PARAMETERS: {"command": "list-projects", "context": "property"}
---
I'll list all your projects in the property context.
```

**API**: Executes `task_project_run list --context "property"`

**Result**: Returns project list to user

## Example: Adding a Project with Checklist

**User**: "Create a new sawmill project in property context with goal to make it operational. Add checklist items for leveling rails, testing cuts, and building an outfeed table."

**LLM Response 1**:
```
EXECUTE_PROGRAM: task-manager
PARAMETERS: {
  "command": "add-project",
  "context": "property",
  "project_name": "Sawmill",
  "goal": "Make sawmill fully operational"
}
---
I'll create the Sawmill project in the property context.
```

**LLM Response 2** (after getting project ID from output):
```
EXECUTE_PROGRAM: task-manager
PARAMETERS: {
  "command": "add-checklist",
  "project_id": 1,
  "items": "Level the carriage rails, Test cuts on scrap wood, Build outfeed table"
}
---
I've added three checklist items to your Sawmill project.
```

## Example: Getting Work Suggestions

**User**: "I'm at the property for 3 days, what should I work on?"

**LLM**:
```
EXECUTE_PROGRAM: task-manager
PARAMETERS: {
  "command": "query-suggest",
  "context": "property",
  "time_available": "3 days"
}
---
Let me check what projects need attention in the property context.
```

**API**: Executes `task_query_run suggest --context "property" --time "3 days"`

**Result**: Returns JSON with scored suggestions based on completion momentum and staleness

## Deployment

### To deploy to your remote server:

1. **Push the changes**:
   ```bash
   cd /home/battlestag/Work/BATTLESTAG-BOT
   git add ORCHESTRATOR/agent-gateway/programs/task_manager.go
   git add ORCHESTRATOR/agent-gateway/main.go
   git commit -m "Add task-manager program to API"
   git push
   ```

2. **On the remote server**:
   ```bash
   cd ~/BATTLESTAG-BOT/ORCHESTRATOR/agent-gateway
   git pull
   go build -o agent-gateway .
   # Restart your agent-gateway service
   sudo systemctl restart agent-gateway  # or however you run it
   ```

3. **Verify**:
   - Check logs: `sudo journalctl -u agent-gateway -f`
   - Should see: "Registered program: task-manager"
   - Should see: "Programs registry initialized (2 programs)" (echo + task-manager)

## Testing

### Via API endpoint:
```bash
curl -X GET http://localhost:8080/api/programs/list
```

Should return JSON including task-manager with all its parameters.

### Via LLM interaction:
```bash
curl -X POST http://localhost:8080/api/llm/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "List my project contexts",
    "session_id": "test-session"
  }'
```

The LLM should automatically use the task-manager program.

## Architecture Notes

### How it Works:
1. **Discovery**: `buildProgramsContext()` in `handlers/llm.go` injects program metadata into every LLM prompt
2. **Invocation**: LLM responds with `EXECUTE_PROGRAM: task-manager` marker
3. **Parsing**: `parseProgramInvocation()` extracts program ID and parameters
4. **Execution**: Registry calls `TaskManagerProgram.Execute()`
5. **Command**: Adapter translates to shell command (e.g., `task_project_run list`)
6. **Result**: Output returned to LLM, which formats it for the user

### No Duplication:
- The `programs/task_manager.go` file does NOT reimplement task management
- It's just a thin wrapper that calls your installed executables
- Your actual task-manager program (in PROGRAMS/task-manager) does all the work
- This keeps the API and the CLI in sync automatically

### Exit Codes:
The wrapper respects your exit codes:
- 0 = Success
- 1 = Invalid arguments
- 2 = Database error
- 3 = Not found

## Benefits

1. **Automatic Discovery**: LLM learns about task-manager capabilities on startup
2. **Natural Language**: Users can request task operations conversationally
3. **Context Aware**: LLM can intelligently suggest what to work on
4. **No Duplication**: Uses your existing executables
5. **Consistent**: API and CLI use the same underlying programs
6. **Extensible**: Easy to add new commands by adding cases to the switch statement

## Future Enhancements

Potential improvements:
- Add `archive-project` command
- Add `remove-checklist-item` command
- Add `reorder-checklist-item` command
- Batch operations (check multiple items at once)
- Project search/filtering
- Export/import functionality

To add a new command:
1. Add a new case to the switch statement in `Execute()`
2. Build the appropriate command args
3. The LLM will automatically see it in the next restart
