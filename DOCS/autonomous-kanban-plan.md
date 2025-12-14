# Autonomous AI Kanban Workflow System

## Vision

An AI agent (Claude) that autonomously manages and executes tasks from a Kanban board:
1. Reads tasks from JSON kanban board
2. Picks tasks from "pull queue"
3. Moves to "WIP" (work in progress)
4. Executes the task using OpenCode
5. Moves completed tasks to "peer review"
6. Logs all actions

## Why This Will Work

We've proven all the components:
✅ Claude can read files autonomously (just tested!)
✅ Claude can analyze and make decisions (objectives test!)
✅ Claude can execute code with opencode (Rails, Go projects!)
✅ We can automate with bash scripts (project-gen, opencode-agent!)

## System Design

### 1. Kanban JSON Structure

```json
{
  "metadata": {
    "version": "1.0",
    "last_updated": "2025-12-06T05:00:00Z",
    "agent": "claude-sonnet-4-5"
  },
  "lanes": {
    "backlog": {
      "name": "Backlog",
      "tasks": []
    },
    "pull_queue": {
      "name": "Pull Queue (Ready to Work)",
      "tasks": [
        {
          "id": "TASK-001",
          "title": "Create disk monitoring script",
          "description": "Build a Python script that monitors disk usage and logs to a file when usage > 80%",
          "priority": "high",
          "complexity": "low",
          "estimated_minutes": 15,
          "tags": ["python", "monitoring", "automation"],
          "acceptance_criteria": [
            "Script checks disk usage of / and /home",
            "Logs to ~/logs/disk-monitor.log",
            "Includes timestamp and percentage",
            "Has error handling"
          ],
          "created": "2025-12-06T04:00:00Z"
        },
        {
          "id": "TASK-002",
          "title": "Build backup script",
          "description": "Create bash script to backup BATTLESTAG-BOT repo to /tmp/backup with timestamp",
          "priority": "high",
          "complexity": "low",
          "estimated_minutes": 10,
          "tags": ["bash", "backup", "automation"],
          "acceptance_criteria": [
            "Creates timestamped backup directory",
            "Copies all files from ~/projects/BATTLESTAG-BOT",
            "Excludes .git directory",
            "Logs success/failure"
          ],
          "created": "2025-12-06T04:00:00Z"
        }
      ]
    },
    "wip": {
      "name": "Work In Progress",
      "limit": 1,
      "tasks": []
    },
    "peer_review": {
      "name": "Peer Review",
      "tasks": []
    },
    "done": {
      "name": "Done",
      "tasks": []
    }
  },
  "history": []
}
```

### 2. Autonomous Agent Script

```bash
#!/bin/bash
# ~/scripts/autonomous-kanban-agent.sh
# Autonomous AI agent that processes Kanban tasks

set -e

KANBAN_FILE="$HOME/kanban-board.json"
OPENCODE_BIN="$HOME/.opencode/bin/opencode"
MODEL="anthropic/claude-sonnet-4-5"
LOG_FILE="$HOME/logs/kanban-agent.log"
WORK_DIR="$HOME/autonomous-work"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

print_header() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}  Autonomous AI Kanban Agent                                ${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
}

# Get next task from pull queue
get_next_task() {
    ruby << 'RUBY'
require 'json'

kanban_file = File.expand_path("~/kanban-board.json")
kanban = JSON.parse(File.read(kanban_file))

pull_queue = kanban["lanes"]["pull_queue"]["tasks"]

if pull_queue.empty?
  puts "NO_TASKS"
  exit 0
end

# Get highest priority task (high > medium > low, then by complexity)
priority_map = { "high" => 0, "medium" => 1, "low" => 2 }
task = pull_queue.min_by { |t| [priority_map[t["priority"]], t["complexity"]] }

puts task.to_json
RUBY
}

# Move task to lane
move_task() {
    local task_id=$1
    local from_lane=$2
    local to_lane=$3

    ruby << RUBY
require 'json'
require 'time'

kanban_file = File.expand_path("~/kanban-board.json")
kanban = JSON.parse(File.read(kanban_file))

# Find and remove task from source lane
task = nil
kanban["lanes"]["#{from_lane}"]["tasks"].reject! do |t|
  if t["id"] == "#{task_id}"
    task = t
    true
  else
    false
  end
end

if task
  # Add to destination lane
  task["moved_to_#{to_lane}"] = Time.now.iso8601
  kanban["lanes"]["#{to_lane}"]["tasks"] << task

  # Add to history
  kanban["history"] << {
    "timestamp" => Time.now.iso8601,
    "action" => "moved",
    "task_id" => "#{task_id}",
    "from" => "#{from_lane}",
    "to" => "#{to_lane}"
  }

  kanban["metadata"]["last_updated"] = Time.now.iso8601

  File.write(kanban_file, JSON.pretty_generate(kanban))
  puts "Task #{task_id} moved from #{from_lane} to #{to_lane}"
else
  puts "Task not found"
  exit 1
end
RUBY
}

# Main execution loop
main() {
    print_header
    log "Starting autonomous agent"

    mkdir -p "$WORK_DIR"
    mkdir -p "$(dirname "$LOG_FILE")"

    # Get next task
    log "Checking pull queue for tasks..."
    TASK_JSON=$(get_next_task)

    if [[ "$TASK_JSON" == "NO_TASKS" ]]; then
        log "No tasks in pull queue. Exiting."
        exit 0
    fi

    TASK_ID=$(echo "$TASK_JSON" | ruby -rjson -e 'puts JSON.parse(STDIN.read)["id"]')
    TASK_TITLE=$(echo "$TASK_JSON" | ruby -rjson -e 'puts JSON.parse(STDIN.read)["title"]')

    log "Selected task: $TASK_ID - $TASK_TITLE"

    # Move to WIP
    log "Moving $TASK_ID to WIP..."
    move_task "$TASK_ID" "pull_queue" "wip"

    # Create work directory for this task
    TASK_WORK_DIR="$WORK_DIR/$TASK_ID"
    mkdir -p "$TASK_WORK_DIR"
    cd "$TASK_WORK_DIR"

    # Execute task with AI
    log "Executing task with Claude..."

    PROMPT=$(cat << EOF
You are an autonomous AI agent executing a task from the Kanban board.

TASK: $TASK_JSON

YOUR MISSION:
1. Analyze the task requirements
2. Implement the solution
3. Test it works
4. Create any necessary documentation

CONSTRAINTS:
- Work in directory: $TASK_WORK_DIR
- Production-ready code with error handling
- Include logging where appropriate
- Follow best practices

Execute this task completely and report when done.
EOF
)

    if $OPENCODE_BIN run --model "$MODEL" --agent=build "$PROMPT"; then
        log "Task $TASK_ID completed successfully"

        # Move to peer review
        log "Moving $TASK_ID to peer review..."
        move_task "$TASK_ID" "wip" "peer_review"

        log "✓ Task $TASK_ID ready for review at: $TASK_WORK_DIR"
    else
        log "✗ Task $TASK_ID failed"
        # Move back to pull queue with failure note
        move_task "$TASK_ID" "wip" "pull_queue"
    fi
}

main "$@"
```

### 3. Continuous Agent Loop (Optional)

```bash
#!/bin/bash
# ~/scripts/autonomous-kanban-loop.sh
# Continuously process Kanban tasks

while true; do
    echo "=== Autonomous Agent Cycle ==="

    # Run the agent
    ~/scripts/autonomous-kanban-agent.sh

    # Check if there are more tasks
    TASKS_REMAINING=$(ruby -rjson -e "
kanban = JSON.parse(File.read(File.expand_path('~/kanban-board.json')))
puts kanban['lanes']['pull_queue']['tasks'].length
")

    if [[ "$TASKS_REMAINING" -eq 0 ]]; then
        echo "No more tasks. Sleeping for 60 seconds..."
        sleep 60
    else
        echo "More tasks available. Processing next..."
        sleep 5
    fi
done
```

### 4. Monitoring Dashboard Script

```bash
#!/bin/bash
# ~/scripts/kanban-status.sh
# Show current Kanban status

ruby << 'RUBY'
require 'json'

kanban_file = File.expand_path("~/kanban-board.json")
kanban = JSON.parse(File.read(kanban_file))

puts "\n╔════════════════════════════════════════════════════════════╗"
puts "║  Autonomous AI Kanban Board Status                        ║"
puts "╚════════════════════════════════════════════════════════════╝\n"

kanban["lanes"].each do |lane_id, lane|
  count = lane["tasks"].length
  puts "#{lane['name']}: #{count} tasks"

  if count > 0
    lane["tasks"].each do |task|
      puts "  • #{task['id']}: #{task['title']} [#{task['priority']}]"
    end
  end
end

puts "\nLast updated: #{kanban['metadata']['last_updated']}"
puts "Total history: #{kanban['history'].length} actions\n"
RUBY
```

## Testing Plan for Tomorrow

### Phase 1: Setup (10 minutes)
1. Create kanban-board.json with 3-4 simple tasks
2. Install the autonomous agent scripts
3. Create logs directory

### Phase 2: Single Task Test (20 minutes)
1. Run agent manually: `~/scripts/autonomous-kanban-agent.sh`
2. Watch it:
   - Read kanban board
   - Pick first task
   - Move to WIP
   - Execute with Claude
   - Move to peer review
3. Verify task completion
4. Check logs

### Phase 3: Multiple Tasks (30 minutes)
1. Add 3 more tasks to pull queue
2. Run agent multiple times
3. Verify all tasks get processed
4. Review generated code/scripts

### Phase 4: Continuous Loop (optional)
1. Start the loop: `~/scripts/autonomous-kanban-loop.sh`
2. Watch it process tasks continuously
3. Add new tasks while running
4. Monitor with: `watch -n 5 ~/scripts/kanban-status.sh`

## Sample Tasks for Testing

### Easy Tasks (Start Here)
```json
{
  "id": "TASK-001",
  "title": "Create hello world script",
  "description": "Make a Python script that prints 'Hello from Autonomous Agent'",
  "priority": "high",
  "complexity": "low",
  "estimated_minutes": 5
}
```

### Medium Tasks
```json
{
  "id": "TASK-002",
  "title": "Disk usage monitor",
  "description": "Python script to check disk usage and log if > 80%",
  "priority": "high",
  "complexity": "medium",
  "estimated_minutes": 15
}
```

### Complex Tasks
```json
{
  "id": "TASK-003",
  "title": "REST API health checker",
  "description": "Go program that checks if battlestag services are running and returns JSON status",
  "priority": "medium",
  "complexity": "high",
  "estimated_minutes": 30
}
```

## Expected Outcomes

After successful testing, you'll have:
✅ Autonomous AI agent managing its own workflow
✅ Self-updating Kanban board in JSON
✅ Complete logs of all actions
✅ Generated code from AI executing tasks
✅ Framework for scaling to complex projects

## Future Enhancements

1. **Priority Intelligence**: AI learns which tasks to prioritize based on system state
2. **Dependency Management**: Tasks can depend on other tasks
3. **Quality Gates**: Auto-run tests before moving to peer review
4. **Notifications**: Slack/email when tasks complete
5. **Multi-Agent**: Multiple AI agents working in parallel
6. **Learning**: Agent improves over time based on success/failure
7. **Integration**: Connect to GitHub Issues, Jira, etc.

## Risk Mitigation

- WIP limit prevents too many concurrent tasks
- All changes logged to history
- Failed tasks go back to pull queue
- Agent runs in isolated work directories
- Can pause/stop loop at any time

## Success Criteria

Tomorrow's test succeeds if:
1. Agent reads kanban board ✓
2. Picks appropriate task ✓
3. Moves task through lanes ✓
4. Executes task with Claude ✓
5. Generates working code ✓
6. Updates JSON correctly ✓

---

**This is the future of autonomous development!** 🚀

The AI becomes your automated junior dev:
- Picks tasks
- Implements them
- Submits for review
- Works 24/7

**Tomorrow we make this real!**
