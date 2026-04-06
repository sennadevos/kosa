# Data Model

## Overview

The system uses 8 core tables. Every table is accessible via the database UI for visual management and via REST API for automation.

```
Accounts ──── Transactions ──── Tags (many-to-many)
                  │                    │
                  │ (loan link)        │
                  │                    │
Loans ──── LoanPayments     Categories (many-to-one)

RecurringRules ──── (generates virtual projections at query time)

BalanceSnapshots ──── Accounts
```

## Tables

### 1. Accounts

Anywhere money lives — bank accounts, savings, investment accounts, credit cards, cash.

| Field      | Type   | Required | Description                                      |
|------------|--------|----------|--------------------------------------------------|
| id         | auto   | auto     | Primary key                                      |
| name       | text   | yes      | "ING Checking", "ABN Savings", "Trading 212"     |
| type       | select | yes      | `checking` / `savings` / `investment` / `credit_card` / `cash` |
| provider   | text   | no       | "ING", "ABN AMRO", "Trading 212", etc.           |
| currency   | text   | no       | EUR (default), USD, etc.                         |
| iban       | text   | no       | IBAN or account number                           |
| is_default | bool   | no       | Primary account for quick-entry commands         |
| notes      | text   | no       |                                                  |

Note: no `balance` field. Balance is derived from BalanceSnapshots + Transactions. See [balance-tracking.md](balance-tracking.md).

**Example accounts:**

| name              | type        | provider    |
|-------------------|-------------|-------------|
| ING Checking      | checking    | ING         |
| ING Joint         | checking    | ING         |
| ABN Savings       | savings     | ABN AMRO    |
| ING Savings       | savings     | ING         |
| Trading 212       | investment  | Trading 212 |
| T212 Card         | credit_card | Trading 212 |
| Cash              | cash        |             |

### 2. Transactions

Every actual money movement. Amounts are always positive — the `type` field determines direction.

| Field            | Type   | Required | Description                                      |
|------------------|--------|----------|--------------------------------------------------|
| id               | auto   | auto     | Primary key                                      |
| date             | date   | yes      | When it happened (defaults to today)             |
| type             | select | yes      | `income` / `expense` / `transfer` / `refund`     |
| amount           | number | yes      | Always positive                                  |
| description      | text   | yes      | "Albert Heijn", "Salary", etc.                   |
| category         | link   | no       | → Categories (one per transaction)               |
| tags             | link[] | no       | → Tags (many-to-many)                            |
| account          | link   | no       | → Accounts (defaults to is_default account)      |
| to_account       | link   | no       | → Accounts (for transfers between accounts)      |
| loan             | link   | no       | → Loans (marks this as a loan movement)          |
| recurring_rule   | link   | no       | → RecurringRules (which rule this fulfills)      |
| refund_of        | link   | no       | → Transactions (the original purchase refunded)  |
| cashback         | number | no       | Cashback earned (e.g., T212 card)                |
| reference        | text   | no       | Bank reference number, invoice ID, etc.          |
| foreign_amount   | number | no       | Original amount in foreign currency              |
| foreign_currency | text   | no       | ISO 4217 currency code (e.g., USD, GBP)         |
| exchange_rate    | number | no       | Effective rate: amount / foreign_amount          |
| notes            | text   | no       |                                                  |

**Loan-linked transactions**: If `loan` is set, the transaction represents a loan movement (borrowing or repayment), not real income/expense. These are included in balance calculations but **excluded from income/expense summaries and spending analysis.** The `type` field still indicates money direction: `income` = money received (borrowing), `expense` = money paid (repayment).

**Refund transactions**: `type = refund` with optional `refund_of` linking to the original purchase. Refunds are inflows in balance calculations and are subtracted from expenses in spending summaries. See [proposals/refund-tracking.md](proposals/refund-tracking.md) for details.

**Recurring rule link**: Optional `recurring_rule` links a transaction to the rule it fulfills. Enables delta detection between projected and actual amounts via `kosa reconcile`. The CLI auto-suggests matches based on amount, category, and date.

**Foreign currency**: `foreign_amount`, `foreign_currency`, and `exchange_rate` are optional fields for transactions involving currency conversion. The `amount` field is always in the account's currency (typically EUR). Foreign fields are informational only — they do not affect balance calculations.

**Minimal entry** (quick coffee): date (auto-today), type (`expense`), amount, description — 4 fields.
**Rich entry** (salary): all fields populated including category, account, recurring_rule, reference, notes.

### 3. RecurringRules

Templates for transactions that repeat on a schedule. These are **never materialized as rows** — projections are computed at query time by evaluating active rules against a date range. See [balance-tracking.md](balance-tracking.md) for how projections work.

| Field        | Type   | Required | Description                                      |
|--------------|--------|----------|--------------------------------------------------|
| id           | auto   | auto     | Primary key                                      |
| name         | text   | yes      | "Netflix", "Rent", "Salary"                      |
| type         | select | yes      | `income` / `expense`                             |
| amount       | number | yes      | Always positive                                  |
| category     | link   | no       | → Categories                                     |
| tags         | link[] | no       | → Tags (many-to-many)                            |
| account      | link   | no       | → Accounts (defaults to is_default)              |
| frequency    | select | yes      | `daily` / `weekly` / `biweekly` / `monthly` / `quarterly` / `yearly` |
| day_of_month | number | no       | 1-31 (for monthly), null for others              |
| start_date   | date   | yes      | When this rule starts                            |
| end_date     | date   | no       | When it ends (null = indekosaite)                 |
| is_active    | bool   | no       | Can pause without deleting (default true)        |
| notes        | text   | no       |                                                  |

### 4. Loans

Track money owed in either direction. Interest is computed, not stored. See [loan-tracking.md](loan-tracking.md) for full details on interest models and the replay function.

| Field              | Type   | Required | Description                                              |
|--------------------|--------|----------|----------------------------------------------------------|
| id                 | auto   | auto     |                                                          |
| type               | select | yes      | `payable` / `receivable`                                 |
| counterparty_name  | text   | yes      | Display name of the other party                          |
| counterparty_uri   | text   | no       | External reference (custom API, CardDAV UID, email, etc) |
| description        | text   | yes      | What the loan is for                                     |
| original_amount    | number | yes      | Principal — the amount originally lent/borrowed          |
| currency           | text   | no       | EUR default                                              |
| date_created       | date   | no       | When the loan was made (defaults to today)               |
| due_date           | date   | no       | When full repayment is expected                          |
| interest_type      | select | no       | `none` / `flat` / `periodic`                             |
| interest_rate      | number | no       | Percentage (e.g., 5 = 5%)                                |
| interest_period    | select | no       | `weekly` / `monthly` / `quarterly` / `yearly`            |
| is_settled         | bool   | no       | True when fully paid off                                 |
| notes              | text   | no       |                                                          |

**Computed fields** (derived at query time, never stored):

| Field             | Calculation                                                    |
|-------------------|----------------------------------------------------------------|
| `total_interest`  | Flat: `original × rate%`. Periodic: derived from replay.       |
| `total_owed`      | `original_amount + total_interest`                             |
| `total_paid`      | `sum(loan_payments.amount)`                                    |
| `remaining`       | No interest / flat: `total_owed - total_paid`. Periodic: output of `replay()`. |
| `is_overdue`      | `due_date < today AND NOT is_settled`                          |

### 5. LoanPayments

Individual payments against a loan. Each payment also links to a Transaction for cash flow visibility.

| Field   | Type   | Required | Description                                        |
|---------|--------|----------|----------------------------------------------------|
| id      | auto   | auto     |                                                    |
| loan    | link   | yes      | → Loans                                            |
| date    | date   | no       | Payment date (defaults to today)                   |
| amount  | number | yes      | Amount paid                                        |
| account | link   | no       | → Accounts (which account the money moved through) |
| notes   | text   | no       |                                                    |

When a loan payment is recorded:
- **Payable** → creates an `expense` transaction on the paying account
- **Receivable** → creates an `income` transaction on the receiving account
- The transaction's `loan` field links back to the loan

### 6. BalanceSnapshots

Periodic anchor points for balance tracking. See [balance-tracking.md](balance-tracking.md) for full details.

| Field   | Type   | Required | Description                                    |
|---------|--------|----------|------------------------------------------------|
| id      | auto   | auto     |                                                |
| account | link   | yes      | → Accounts                                     |
| date    | date   | yes      | When the balance was verified                  |
| balance | number | yes      | The verified balance                           |
| source  | select | no       | `manual` / `bank_import` / `reconciliation`    |
| notes   | text   | no       |                                                |

## Balance Calculations

Balance is always derived, never stored on the account. See [balance-tracking.md](balance-tracking.md) for full formulas.

```
current_balance(account) =
    latest_snapshot.balance
  + sum(income since snapshot)
  + sum(refunds since snapshot)
  - sum(expenses since snapshot)
  ± transfers

spending_summary(category, period) =
    sum(expenses in period)
  - sum(refunds in period)
  # excludes loan-linked transactions

projected_balance(account, future_date) =
    current_balance(account)
  + sum(recurring_rule_occurrences(today, future_date))
```

## 7. Categories

Each transaction has at most one category. Categories are the primary axis for spending summaries — totals always sum cleanly because each transaction belongs to exactly one category.

| Field | Type   | Required | Description                          |
|-------|--------|----------|--------------------------------------|
| id    | auto   | auto     | Primary key                          |
| name  | text   | yes      | "groceries", "rent", "salary", etc.  |
| type  | select | no       | `income` / `expense` / `neutral`     |

The dekosaitive list of categories lives in the database. Adding or renaming a category is a single row operation in the UI — no schema change needed.

## 8. Tags

Tags provide flexible, freeform grouping across transactions. A transaction can have zero or many tags. Tags are not used for spending summaries — use categories for that.

| Field | Type   | Required | Description                                    |
|-------|--------|----------|------------------------------------------------|
| id    | auto   | auto     | Primary key                                    |
| name  | text   | yes      | "weekly-shop", "date-night", "tax-deductible"  |

Tags are many-to-many with Transactions and RecurringRules. Use them for filtering, searching, and ad-hoc grouping that doesn't fit into categories.
