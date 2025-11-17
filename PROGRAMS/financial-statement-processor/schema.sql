-- Financial Statement Processor Database Schema
-- SQLite 3

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Account information
    account_name TEXT NOT NULL,
    account_last4 TEXT NOT NULL,

    -- Transaction details
    transaction_date DATE NOT NULL,
    post_date DATE,
    description TEXT NOT NULL,
    amount REAL NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('debit', 'credit')),
    balance REAL,

    -- Statement metadata
    statement_date DATE NOT NULL,
    source_file TEXT,

    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- Prevent duplicate transactions
    UNIQUE(account_last4, transaction_date, description, amount)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_transactions_account
    ON transactions(account_name, account_last4);

CREATE INDEX IF NOT EXISTS idx_transactions_date
    ON transactions(transaction_date);

CREATE INDEX IF NOT EXISTS idx_transactions_post_date
    ON transactions(post_date);

CREATE INDEX IF NOT EXISTS idx_transactions_statement_date
    ON transactions(statement_date);

CREATE INDEX IF NOT EXISTS idx_transactions_type
    ON transactions(transaction_type);

CREATE INDEX IF NOT EXISTS idx_transactions_account_date
    ON transactions(account_name, transaction_date DESC);

-- Processing log table (tracks statement processing history)
CREATE TABLE IF NOT EXISTS processing_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_file TEXT NOT NULL,
    statement_date DATE,
    account_name TEXT,
    transactions_inserted INTEGER DEFAULT 0,
    transactions_skipped INTEGER DEFAULT 0,
    status TEXT NOT NULL CHECK (status IN ('success', 'parse_error', 'db_error')),
    error_message TEXT,
    processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_processing_log_file
    ON processing_log(source_file);

CREATE INDEX IF NOT EXISTS idx_processing_log_status
    ON processing_log(status);

-- Trigger to automatically update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_transactions_updated_at
    AFTER UPDATE ON transactions
    FOR EACH ROW
BEGIN
    UPDATE transactions SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;

-- Note: SQLite doesn't support views in the same way as PostgreSQL,
-- but you can query directly for summaries:
--
-- Account Summary:
--   SELECT account_name, account_last4, COUNT(*) as transaction_count,
--          SUM(CASE WHEN transaction_type = 'debit' THEN amount ELSE 0 END) as total_debits,
--          SUM(CASE WHEN transaction_type = 'credit' THEN amount ELSE 0 END) as total_credits,
--          MIN(transaction_date) as first_transaction,
--          MAX(transaction_date) as last_transaction,
--          MAX(statement_date) as latest_statement
--   FROM transactions
--   GROUP BY account_name, account_last4;
--
-- Monthly Summary:
--   SELECT account_name, account_last4,
--          strftime('%Y-%m', transaction_date) as month,
--          COUNT(*) as transaction_count,
--          SUM(CASE WHEN transaction_type = 'debit' THEN amount ELSE 0 END) as total_debits,
--          SUM(CASE WHEN transaction_type = 'credit' THEN amount ELSE 0 END) as total_credits
--   FROM transactions
--   GROUP BY account_name, account_last4, strftime('%Y-%m', transaction_date)
--   ORDER BY month DESC, account_name;
