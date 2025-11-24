package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// Context represents a top-level context (property, house, pxp, personal)
type Context struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Project represents a tracked project
type Project struct {
	ID          int64     `json:"id"`
	ContextID   int64     `json:"context_id"`
	ContextName string    `json:"context_name,omitempty"`
	Name        string    `json:"name"`
	Goal        string    `json:"goal,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	LastTouched time.Time `json:"last_touched"`
}

// ChecklistItem represents a granular task within a project
type ChecklistItem struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	Description string     `json:"description"`
	Done        bool       `json:"done"`
	SortOrder   int        `json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// JournalEntry represents a narrative log entry
type JournalEntry struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectStatus represents the combined status of a project
type ProjectStatus struct {
	Project          Project           `json:"project"`
	ProgressPercent  int               `json:"progress_percent"`
	ChecklistTotal   int               `json:"checklist_total"`
	ChecklistDone    int               `json:"checklist_done"`
	DaysSinceTouched int               `json:"days_since_touched"`
	RecentJournal    []*JournalEntry   `json:"recent_journal"`
	PendingItems     []*ChecklistItem  `json:"pending_items"`
}

// New creates a new database connection and initializes schema
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Enable foreign keys
	_, err = conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema creates the necessary tables if they don't exist
func (db *DB) initSchema() error {
	schema := `
	-- Contexts (property, house, pxp, personal)
	CREATE TABLE IF NOT EXISTS contexts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Projects
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		context_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		goal TEXT,
		status TEXT DEFAULT 'active' CHECK(status IN ('active', 'paused', 'blocked', 'complete')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_touched DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (context_id) REFERENCES contexts(id)
	);

	-- Checklist items (granular tasks within projects)
	CREATE TABLE IF NOT EXISTS checklist_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		description TEXT NOT NULL,
		done BOOLEAN DEFAULT 0,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		FOREIGN KEY (project_id) REFERENCES projects(id)
	);

	-- Journal entries (narrative log)
	CREATE TABLE IF NOT EXISTS journal_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_projects_context ON projects(context_id);
	CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
	CREATE INDEX IF NOT EXISTS idx_projects_last_touched ON projects(last_touched);
	CREATE INDEX IF NOT EXISTS idx_checklist_project ON checklist_items(project_id);
	CREATE INDEX IF NOT EXISTS idx_checklist_done ON checklist_items(done);
	CREATE INDEX IF NOT EXISTS idx_journal_project ON journal_entries(project_id);
	CREATE INDEX IF NOT EXISTS idx_journal_created ON journal_entries(created_at);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	return nil
}

// ========== Context Methods ==========

// AddContext adds a new context
func (db *DB) AddContext(name string) (int64, error) {
	query := `INSERT INTO contexts (name) VALUES (?)`
	result, err := db.conn.Exec(query, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, fmt.Errorf("context already exists")
		}
		return 0, fmt.Errorf("insert context: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return id, nil
}

// ListContexts returns all contexts
func (db *DB) ListContexts() ([]*Context, error) {
	query := `SELECT id, name, created_at FROM contexts ORDER BY name`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query contexts: %w", err)
	}
	defer rows.Close()

	var contexts []*Context
	for rows.Next() {
		ctx := &Context{}
		if err := rows.Scan(&ctx.ID, &ctx.Name, &ctx.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan context: %w", err)
		}
		contexts = append(contexts, ctx)
	}

	return contexts, rows.Err()
}

// GetContextByName retrieves a context by name
func (db *DB) GetContextByName(name string) (*Context, error) {
	query := `SELECT id, name, created_at FROM contexts WHERE name = ?`
	ctx := &Context{}
	err := db.conn.QueryRow(query, name).Scan(&ctx.ID, &ctx.Name, &ctx.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("context not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query context: %w", err)
	}
	return ctx, nil
}

// ========== Project Methods ==========

// AddProject adds a new project
func (db *DB) AddProject(contextID int64, name, goal string) (int64, error) {
	now := time.Now()
	query := `
		INSERT INTO projects (context_id, name, goal, created_at, last_touched)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := db.conn.Exec(query, contextID, name, goal, now, now)
	if err != nil {
		return 0, fmt.Errorf("insert project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return id, nil
}

// UpdateProject updates a project's status or goal
func (db *DB) UpdateProject(id int64, status, goal *string) error {
	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}

	if status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *status)
	}

	if goal != nil {
		updates = append(updates, "goal = ?")
		args = append(args, *goal)
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates specified")
	}

	// Always update last_touched
	updates = append(updates, "last_touched = ?")
	args = append(args, time.Now())

	// Add ID to args
	args = append(args, id)

	query := fmt.Sprintf("UPDATE projects SET %s WHERE id = ?", strings.Join(updates, ", "))
	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// ArchiveProject marks a project as complete
func (db *DB) ArchiveProject(id int64) error {
	status := "complete"
	return db.UpdateProject(id, &status, nil)
}

// GetProject retrieves a project by ID
func (db *DB) GetProject(id int64) (*Project, error) {
	query := `
		SELECT p.id, p.context_id, c.name as context_name, p.name, p.goal, p.status,
		       p.created_at, p.last_touched
		FROM projects p
		JOIN contexts c ON p.context_id = c.id
		WHERE p.id = ?
	`

	proj := &Project{}
	err := db.conn.QueryRow(query, id).Scan(
		&proj.ID, &proj.ContextID, &proj.ContextName, &proj.Name, &proj.Goal,
		&proj.Status, &proj.CreatedAt, &proj.LastTouched,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query project: %w", err)
	}

	return proj, nil
}

// GetProjectByName retrieves a project by name
func (db *DB) GetProjectByName(name string) (*Project, error) {
	query := `
		SELECT p.id, p.context_id, c.name as context_name, p.name, p.goal, p.status,
		       p.created_at, p.last_touched
		FROM projects p
		JOIN contexts c ON p.context_id = c.id
		WHERE p.name = ?
	`

	proj := &Project{}
	err := db.conn.QueryRow(query, name).Scan(
		&proj.ID, &proj.ContextID, &proj.ContextName, &proj.Name, &proj.Goal,
		&proj.Status, &proj.CreatedAt, &proj.LastTouched,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query project: %w", err)
	}

	return proj, nil
}

// ListProjects retrieves projects with optional filters
func (db *DB) ListProjects(contextName, status string) ([]*Project, error) {
	query := `
		SELECT p.id, p.context_id, c.name as context_name, p.name, p.goal, p.status,
		       p.created_at, p.last_touched
		FROM projects p
		JOIN contexts c ON p.context_id = c.id
		WHERE 1=1
	`

	args := []interface{}{}

	if contextName != "" {
		query += " AND c.name = ?"
		args = append(args, contextName)
	}

	if status != "" {
		query += " AND p.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY p.last_touched DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		proj := &Project{}
		err := rows.Scan(
			&proj.ID, &proj.ContextID, &proj.ContextName, &proj.Name, &proj.Goal,
			&proj.Status, &proj.CreatedAt, &proj.LastTouched,
		)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, proj)
	}

	return projects, rows.Err()
}

// TouchProject updates the last_touched timestamp
func (db *DB) TouchProject(projectID int64) error {
	query := `UPDATE projects SET last_touched = ? WHERE id = ?`
	result, err := db.conn.Exec(query, time.Now(), projectID)
	if err != nil {
		return fmt.Errorf("touch project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// ========== Checklist Methods ==========

// AddChecklistItem adds a new checklist item
func (db *DB) AddChecklistItem(projectID int64, description string) (int64, error) {
	// Get max sort order for this project
	var maxOrder int
	err := db.conn.QueryRow(`
		SELECT COALESCE(MAX(sort_order), -1) FROM checklist_items WHERE project_id = ?
	`, projectID).Scan(&maxOrder)
	if err != nil {
		return 0, fmt.Errorf("get max sort order: %w", err)
	}

	query := `
		INSERT INTO checklist_items (project_id, description, sort_order)
		VALUES (?, ?, ?)
	`
	result, err := db.conn.Exec(query, projectID, description, maxOrder+1)
	if err != nil {
		return 0, fmt.Errorf("insert checklist item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	// Touch the project
	if err := db.TouchProject(projectID); err != nil {
		return id, fmt.Errorf("touch project: %w", err)
	}

	return id, nil
}

// CheckItem marks a checklist item as complete
func (db *DB) CheckItem(id int64) error {
	// Get project ID first
	var projectID int64
	err := db.conn.QueryRow(`SELECT project_id FROM checklist_items WHERE id = ?`, id).Scan(&projectID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("item not found")
	}
	if err != nil {
		return fmt.Errorf("get project id: %w", err)
	}

	query := `UPDATE checklist_items SET done = 1, completed_at = ? WHERE id = ?`
	result, err := db.conn.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("check item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("item not found")
	}

	// Touch the project
	if err := db.TouchProject(projectID); err != nil {
		return fmt.Errorf("touch project: %w", err)
	}

	return nil
}

// UncheckItem marks a checklist item as incomplete
func (db *DB) UncheckItem(id int64) error {
	// Get project ID first
	var projectID int64
	err := db.conn.QueryRow(`SELECT project_id FROM checklist_items WHERE id = ?`, id).Scan(&projectID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("item not found")
	}
	if err != nil {
		return fmt.Errorf("get project id: %w", err)
	}

	query := `UPDATE checklist_items SET done = 0, completed_at = NULL WHERE id = ?`
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("uncheck item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("item not found")
	}

	// Touch the project
	if err := db.TouchProject(projectID); err != nil {
		return fmt.Errorf("touch project: %w", err)
	}

	return nil
}

// RemoveChecklistItem removes a checklist item
func (db *DB) RemoveChecklistItem(id int64) error {
	query := `DELETE FROM checklist_items WHERE id = ?`
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("remove item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("item not found")
	}

	return nil
}

// ReorderChecklistItem changes the sort order of an item
func (db *DB) ReorderChecklistItem(id int64, position int) error {
	query := `UPDATE checklist_items SET sort_order = ? WHERE id = ?`
	result, err := db.conn.Exec(query, position, id)
	if err != nil {
		return fmt.Errorf("reorder item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("item not found")
	}

	return nil
}

// ListChecklistItems retrieves checklist items for a project
func (db *DB) ListChecklistItems(projectID int64, pendingOnly bool) ([]*ChecklistItem, error) {
	query := `
		SELECT id, project_id, description, done, sort_order, created_at, completed_at
		FROM checklist_items
		WHERE project_id = ?
	`

	if pendingOnly {
		query += " AND done = 0"
	}

	query += " ORDER BY sort_order, created_at"

	rows, err := db.conn.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("query checklist items: %w", err)
	}
	defer rows.Close()

	var items []*ChecklistItem
	for rows.Next() {
		item := &ChecklistItem{}
		err := rows.Scan(
			&item.ID, &item.ProjectID, &item.Description, &item.Done,
			&item.SortOrder, &item.CreatedAt, &item.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan checklist item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetChecklistStats returns stats for a project's checklist
func (db *DB) GetChecklistStats(projectID int64) (total, done int, err error) {
	err = db.conn.QueryRow(`
		SELECT COUNT(*), SUM(CASE WHEN done = 1 THEN 1 ELSE 0 END)
		FROM checklist_items WHERE project_id = ?
	`, projectID).Scan(&total, &done)
	return
}

// ========== Journal Methods ==========

// AddJournalEntry adds a new journal entry
func (db *DB) AddJournalEntry(projectID int64, content string) (int64, error) {
	query := `INSERT INTO journal_entries (project_id, content) VALUES (?, ?)`
	result, err := db.conn.Exec(query, projectID, content)
	if err != nil {
		return 0, fmt.Errorf("insert journal entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	// Touch the project
	if err := db.TouchProject(projectID); err != nil {
		return id, fmt.Errorf("touch project: %w", err)
	}

	return id, nil
}

// ListJournalEntries retrieves journal entries for a project
func (db *DB) ListJournalEntries(projectID int64, limit int) ([]*JournalEntry, error) {
	query := `
		SELECT id, project_id, content, created_at
		FROM journal_entries
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.conn.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("query journal entries: %w", err)
	}
	defer rows.Close()

	var entries []*JournalEntry
	for rows.Next() {
		entry := &JournalEntry{}
		err := rows.Scan(&entry.ID, &entry.ProjectID, &entry.Content, &entry.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan journal entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// GetLatestJournalEntry retrieves the most recent journal entry for a project
func (db *DB) GetLatestJournalEntry(projectID int64) (*JournalEntry, error) {
	query := `
		SELECT id, project_id, content, created_at
		FROM journal_entries
		WHERE project_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	entry := &JournalEntry{}
	err := db.conn.QueryRow(query, projectID).Scan(
		&entry.ID, &entry.ProjectID, &entry.Content, &entry.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no journal entries found")
	}
	if err != nil {
		return nil, fmt.Errorf("query journal entry: %w", err)
	}

	return entry, nil
}

// ========== Query Methods ==========

// GetProjectStatus returns comprehensive status for a project
func (db *DB) GetProjectStatus(projectID int64) (*ProjectStatus, error) {
	proj, err := db.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	status := &ProjectStatus{
		Project: *proj,
	}

	// Get checklist stats
	total, done, err := db.GetChecklistStats(projectID)
	if err != nil {
		return nil, fmt.Errorf("get checklist stats: %w", err)
	}
	status.ChecklistTotal = total
	status.ChecklistDone = done

	// Calculate progress
	if total > 0 {
		status.ProgressPercent = (done * 100) / total
	}

	// Calculate days since touched
	status.DaysSinceTouched = int(time.Since(proj.LastTouched).Hours() / 24)

	// Get recent journal entries (last 3)
	journal, err := db.ListJournalEntries(projectID, 3)
	if err != nil {
		return nil, fmt.Errorf("get journal entries: %w", err)
	}
	status.RecentJournal = journal

	// Get pending items
	pending, err := db.ListChecklistItems(projectID, true)
	if err != nil {
		return nil, fmt.Errorf("get pending items: %w", err)
	}
	status.PendingItems = pending

	return status, nil
}

// GetNeglectedProjects returns projects not touched in specified days
func (db *DB) GetNeglectedProjects(days int) ([]*Project, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	query := `
		SELECT p.id, p.context_id, c.name as context_name, p.name, p.goal, p.status,
		       p.created_at, p.last_touched
		FROM projects p
		JOIN contexts c ON p.context_id = c.id
		WHERE p.last_touched < ? AND p.status = 'active'
		ORDER BY p.last_touched ASC
	`

	rows, err := db.conn.Query(query, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query neglected projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		proj := &Project{}
		err := rows.Scan(
			&proj.ID, &proj.ContextID, &proj.ContextName, &proj.Name, &proj.Goal,
			&proj.Status, &proj.CreatedAt, &proj.LastTouched,
		)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, proj)
	}

	return projects, rows.Err()
}

// GetProjectsNearCompletion returns projects above threshold% complete
func (db *DB) GetProjectsNearCompletion(threshold int) ([]*ProjectStatus, error) {
	projects, err := db.ListProjects("", "active")
	if err != nil {
		return nil, err
	}

	var nearComplete []*ProjectStatus
	for _, proj := range projects {
		total, done, err := db.GetChecklistStats(proj.ID)
		if err != nil {
			continue
		}

		if total == 0 {
			continue
		}

		progress := (done * 100) / total
		if progress >= threshold {
			status := &ProjectStatus{
				Project:         *proj,
				ProgressPercent: progress,
				ChecklistTotal:  total,
				ChecklistDone:   done,
			}
			nearComplete = append(nearComplete, status)
		}
	}

	return nearComplete, nil
}

// GetContextProjects returns all projects in a context with their stats
func (db *DB) GetContextProjects(contextName string) ([]*ProjectStatus, error) {
	projects, err := db.ListProjects(contextName, "")
	if err != nil {
		return nil, err
	}

	var statuses []*ProjectStatus
	for _, proj := range projects {
		total, done, err := db.GetChecklistStats(proj.ID)
		if err != nil {
			continue
		}

		progress := 0
		if total > 0 {
			progress = (done * 100) / total
		}

		daysSince := int(time.Since(proj.LastTouched).Hours() / 24)

		status := &ProjectStatus{
			Project:          *proj,
			ProgressPercent:  progress,
			ChecklistTotal:   total,
			ChecklistDone:    done,
			DaysSinceTouched: daysSince,
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}
