# Architecture

## Overview

```
┌──────────────────────────────────────────────────┐
│                   Consumers                       │
│                                                   │
│   CLI (Go)     Claude Agents     Cron / Triggers  │
└──────┬───────────────┬───────────────┬────────────┘
       │               │               │
       ▼               ▼               ▼
┌──────────────────────────────────────────────────┐
│              Application Layer                    │
│                                                   │
│   Commands      Queries      Computed Fields      │
│   (spend,       (balance,    (loan replay,        │
│    split,        summary,     projections,         │
│    refund)       reconcile)   reconciliation)      │
└──────────────────────┬───────────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────────┐
│              Backend Interface                    │
│                                                   │
│   A Go interface that abstracts all database      │
│   operations. The application layer calls this    │
│   interface — it never talks to a specific        │
│   database or API directly.                       │
└──────────────────────┬───────────────────────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
     ┌─────────┐ ┌──────────┐ ┌──────────┐
     │ Teable  │ │ Directus │ │ Postgres │
     │ Backend │ │ Backend  │ │ Backend  │
     │ (REST)  │ │ (future) │ │ (future) │
     └─────────┘ └──────────┘ └──────────┘
          │            │            │
          ▼            ▼            ▼
     ┌──────────────────────────────────┐
     │          PostgreSQL              │
     └──────────────────────────────────┘
```

All backends ultimately read/write the same PostgreSQL database. The difference is how they access it — via Teable's REST API, Directus's REST/GraphQL API, or raw SQL.

## Backend Interface

The interface dekosaes CRUD operations for each domain entity plus query methods for computed results. The application layer depends only on this interface.

```go
type Backend interface {
    // Accounts
    ListAccounts(ctx context.Context, opts AccountFilter) ([]Account, error)
    GetAccount(ctx context.Context, id string) (*Account, error)
    CreateAccount(ctx context.Context, a AccountInput) (*Account, error)
    UpdateAccount(ctx context.Context, id string, a AccountInput) (*Account, error)

    // Transactions
    ListTransactions(ctx context.Context, opts TransactionFilter) ([]Transaction, error)
    GetTransaction(ctx context.Context, id string) (*Transaction, error)
    CreateTransaction(ctx context.Context, t TransactionInput) (*Transaction, error)
    UpdateTransaction(ctx context.Context, id string, t TransactionInput) (*Transaction, error)
    DeleteTransaction(ctx context.Context, id string) error

    // Recurring Rules
    ListRecurringRules(ctx context.Context, opts RecurringRuleFilter) ([]RecurringRule, error)
    CreateRecurringRule(ctx context.Context, r RecurringRuleInput) (*RecurringRule, error)
    UpdateRecurringRule(ctx context.Context, id string, r RecurringRuleInput) (*RecurringRule, error)

    // Loans
    ListLoans(ctx context.Context, opts LoanFilter) ([]Loan, error)
    GetLoan(ctx context.Context, id string) (*Loan, error)
    CreateLoan(ctx context.Context, l LoanInput) (*Loan, error)
    UpdateLoan(ctx context.Context, id string, l LoanInput) (*Loan, error)

    // Loan Payments
    ListLoanPayments(ctx context.Context, loanID string) ([]LoanPayment, error)
    CreateLoanPayment(ctx context.Context, p LoanPaymentInput) (*LoanPayment, error)

    // Balance Snapshots
    ListSnapshots(ctx context.Context, accountID string) ([]BalanceSnapshot, error)
    LatestSnapshot(ctx context.Context, accountID string) (*BalanceSnapshot, error)
    CreateSnapshot(ctx context.Context, s SnapshotInput) (*BalanceSnapshot, error)
}
```

Each method maps to one or more API calls to the underlying backend. The interface uses domain types (`Transaction`, `Loan`, etc.), not API-specific types — each backend translates between its wire format and the domain types.

## Application Layer

The application layer contains all business logic. It depends only on the `Backend` interface, never on a specific backend implementation.

### Commands

Commands mutate state. Each command validates input, calls the backend, and may trigger side effects.

| Command | What it does |
|---------|-------------|
| `Spend` | Creates an expense transaction |
| `Income` | Creates an income transaction |
| `Transfer` | Creates a transfer transaction between two accounts |
| `Refund` | Creates a refund transaction, optionally linked to the original |
| `Split` | Creates personal expense + N loans + N loan-linked expenses |
| `Owe` | Shortcut: creates a payable loan (no interest) |
| `Lent` | Shortcut: creates a receivable loan (no interest) |
| `LoanNew` | Creates a loan with full options (interest, period, due date) |
| `LoanPay` | Creates a loan payment + corresponding transaction |
| `RecurringAdd` | Creates a recurring rule |
| `Snapshot` | Records a balance snapshot |

### Queries

Queries are read-only. Some are simple pass-throughs to the backend, others compute derived values.

| Query | What it computes |
|-------|-----------------|
| `Balance` | Current balance: latest snapshot + transactions since |
| `ProjectedBalance` | Future balance: current + recurring rules applied forward |
| `HistoricalBalance` | Past balance: walk from nearest snapshot |
| `Reconcile` | Compare actual transactions to recurring rule projections, surface deltas |
| `LoanStatus` | Replay loan timeline to compute remaining balance, total interest, total paid |
| `SpendingSummary` | Net expenses - refunds per category, excluding loan-linked transactions |

### Computed Logic

These functions live in the application layer, not the backend:

- **Loan replay**: walks loan timeline to compute periodic interest and remaining balance
- **Projection generation**: evaluates recurring rules against a date range to produce virtual future transactions
- **Reconciliation matching**: pairs actual transactions to recurring rules via the `recurring_rule` link
- **Balance derivation**: combines snapshots + transactions + projections

This is intentional — computed logic must produce identical results regardless of which backend is active.

## Teable Backend

The first (and currently only) backend implementation. Talks to Teable's REST API.

### Configuration

```toml
# ~/.config/kosa/config.toml
[backend]
type = "teable"

[backend.teable]
url = "http://localhost:3000"
token = "your-personal-access-token"

# Table IDs — Teable uses internal IDs, not table names
[backend.teable.tables]
accounts = "tbl_xxxxxxxx"
transactions = "tbl_xxxxxxxx"
recurring_rules = "tbl_xxxxxxxx"
loans = "tbl_xxxxxxxx"
loan_payments = "tbl_xxxxxxxx"
balance_snapshots = "tbl_xxxxxxxx"
```

Table and field IDs are configured once after setting up the Teable space. The backend uses these IDs to construct API calls.

### API Mapping

| Operation | Teable API |
|-----------|-----------|
| List with filters | `GET /api/table/{tableId}/record?filter=...` |
| Get by ID | `GET /api/table/{tableId}/record/{recordId}` |
| Create | `POST /api/table/{tableId}/record` |
| Update | `PATCH /api/table/{tableId}/record/{recordId}` |
| Delete | `DELETE /api/table/{tableId}/record/{recordId}` |

The backend translates between Teable's field IDs and the domain model's field names. This mapping is configured alongside the table IDs.

### Field Mapping

Teable uses internal field IDs (e.g., `fld_xxxxxxxx`). The backend maintains a mapping:

```toml
[backend.teable.fields.transactions]
date = "fld_xxxxxxxx"
type = "fld_xxxxxxxx"
amount = "fld_xxxxxxxx"
description = "fld_xxxxxxxx"
category = "fld_xxxxxxxx"
account = "fld_xxxxxxxx"
# ...
```

This is verbose but explicit. An alternative is to auto-discover field IDs by name on first run.

## Project Structure

```
kosa/
├── cmd/                    # CLI entry points (cobra commands)
│   ├── root.go
│   ├── spend.go
│   ├── income.go
│   ├── transfer.go
│   ├── refund.go
│   ├── split.go
│   ├── loan.go
│   ├── recurring.go
│   ├── balance.go
│   ├── reconcile.go
│   ├── snapshot.go
│   └── list.go
├── internal/
│   ├── domain/             # Domain types (Transaction, Loan, Account, etc.)
│   │   ├── transaction.go
│   │   ├── account.go
│   │   ├── loan.go
│   │   ├── recurring.go
│   │   └── snapshot.go
│   ├── app/                # Application layer (commands + queries)
│   │   ├── spend.go
│   │   ├── split.go
│   │   ├── balance.go
│   │   ├── reconcile.go
│   │   ├── loan_replay.go
│   │   └── projection.go
│   ├── backend/            # Backend interface + implementations
│   │   ├── backend.go      # Interface dekosaition
│   │   ├── teable/         # Teable implementation
│   │   │   ├── client.go
│   │   │   ├── mapping.go
│   │   │   └── teable.go
│   │   ├── directus/       # Future
│   │   └── postgres/       # Future
│   └── config/             # Configuration loading
│       └── config.go
├── go.mod
├── go.sum
└── config.example.toml
```

## Design Principles

1. **Backend interface is the boundary.** Everything above it is backend-agnostic. Everything below it is backend-specific.
2. **Domain types are the lingua franca.** The application layer and CLI work with `Transaction`, `Loan`, etc. — never with API-specific types like Teable records.
3. **Computed logic lives in the application layer.** Loan replay, projections, reconciliation, and balance derivation are pure functions of domain types. They produce identical results regardless of backend.
4. **Configuration over convention.** Table IDs, field IDs, and API URLs are explicit in config. No magic name matching.
5. **Backend implementations are thin.** They translate between wire format and domain types. No business logic in backends.
