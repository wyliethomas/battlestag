-- Financial Statement Processor Database Schema
-- PostgreSQL 12+

-- Create database (run as postgres superuser)
-- CREATE DATABASE financial_data;
-- \c financial_data

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,

    -- Account information
    account_name VARCHAR(255) NOT NULL,
    account_last4 VARCHAR(4) NOT NULL,

    -- Transaction details
    transaction_date DATE NOT NULL,
    post_date DATE,
    description TEXT NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('debit', 'credit')),
    balance DECIMAL(12, 2),

    -- Statement metadata
    statement_date DATE NOT NULL,
    source_file VARCHAR(500),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

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

-- Processing log table (optional, for tracking statement processing)
CREATE TABLE IF NOT EXISTS processing_log (
    id BIGSERIAL PRIMARY KEY,
    source_file VARCHAR(500) NOT NULL,
    statement_date DATE,
    account_name VARCHAR(255),
    transactions_inserted INTEGER DEFAULT 0,
    transactions_skipped INTEGER DEFAULT 0,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'parse_error', 'db_error')),
    error_message TEXT,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_processing_log_file
    ON processing_log(source_file);

CREATE INDEX IF NOT EXISTS idx_processing_log_status
    ON processing_log(status);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at
CREATE TRIGGER update_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- View for quick account summaries
CREATE OR REPLACE VIEW account_summary AS
SELECT
    account_name,
    account_last4,
    COUNT(*) as transaction_count,
    SUM(CASE WHEN transaction_type = 'debit' THEN amount ELSE 0 END) as total_debits,
    SUM(CASE WHEN transaction_type = 'credit' THEN amount ELSE 0 END) as total_credits,
    MIN(transaction_date) as first_transaction,
    MAX(transaction_date) as last_transaction,
    MAX(statement_date) as latest_statement
FROM transactions
GROUP BY account_name, account_last4
ORDER BY account_name;

-- View for monthly summaries
CREATE OR REPLACE VIEW monthly_summary AS
SELECT
    account_name,
    account_last4,
    DATE_TRUNC('month', transaction_date) as month,
    COUNT(*) as transaction_count,
    SUM(CASE WHEN transaction_type = 'debit' THEN amount ELSE 0 END) as total_debits,
    SUM(CASE WHEN transaction_type = 'credit' THEN amount ELSE 0 END) as total_credits,
    SUM(CASE WHEN transaction_type = 'credit' THEN amount ELSE -amount END) as net_change
FROM transactions
GROUP BY account_name, account_last4, DATE_TRUNC('month', transaction_date)
ORDER BY month DESC, account_name;

-- Grant permissions (adjust as needed)
-- GRANT SELECT, INSERT, UPDATE ON transactions TO financial_app;
-- GRANT SELECT, INSERT ON processing_log TO financial_app;
-- GRANT SELECT ON account_summary, monthly_summary TO financial_app;
-- GRANT USAGE, SELECT ON SEQUENCE transactions_id_seq TO financial_app;
-- GRANT USAGE, SELECT ON SEQUENCE processing_log_id_seq TO financial_app;
