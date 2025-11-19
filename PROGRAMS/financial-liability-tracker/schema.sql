-- Financial Liability Tracker Schema
-- SQLite

-- Liabilities table
CREATE TABLE IF NOT EXISTS liabilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    liability_type TEXT NOT NULL CHECK (liability_type IN (
        'credit-card', 'auto-loan', 'mortgage',
        'student-loan', 'personal-loan', 'medical-debt'
    )),
    current_balance REAL NOT NULL,
    original_amount REAL,
    credit_limit REAL,  -- For credit cards
    interest_rate REAL,
    minimum_payment REAL,
    creditor_name TEXT,
    account_last4 TEXT,
    opened_date TEXT,  -- ISO8601 format: YYYY-MM-DD
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Balance history table
CREATE TABLE IF NOT EXISTS liability_balance_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    liability_id INTEGER NOT NULL REFERENCES liabilities(id) ON DELETE CASCADE,
    balance REAL NOT NULL,
    recorded_at TEXT DEFAULT CURRENT_TIMESTAMP,
    notes TEXT
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_liabilities_type ON liabilities(liability_type);
CREATE INDEX IF NOT EXISTS idx_liabilities_name ON liabilities(name);
CREATE INDEX IF NOT EXISTS idx_balance_history_liability_id ON liability_balance_history(liability_id);
CREATE INDEX IF NOT EXISTS idx_balance_history_recorded_at ON liability_balance_history(recorded_at);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_liabilities_updated_at
AFTER UPDATE ON liabilities
FOR EACH ROW
BEGIN
    UPDATE liabilities SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
