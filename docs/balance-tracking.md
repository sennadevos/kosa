# Balance Tracking

## Overview

The system uses a hybrid approach: **balance snapshots** as trusted anchor points, with **transactions** filling the gaps between them. This avoids the all-or-nothing problem — you don't need to log every coffee to have a useful balance, but the more you log the more accurate it gets.

## Components

### Balance Snapshots

A snapshot records a verified balance for an account at a point in time.

| Field   | Type   | Required | Description                                    |
|---------|--------|----------|------------------------------------------------|
| id      | auto   | auto     |                                                |
| account | link   | yes      | Which account                                  |
| date    | date   | yes      | When the balance was verified                  |
| balance | number | yes      | The verified balance                           |
| source  | select | no       | `manual` / `bank_import` / `reconciliation`    |
| notes   | text   | no       |                                                |

Sources of snapshots:
- **Manual**: you open your banking app, type the number in
- **Bank import**: parsed from a CSV/PDF statement
- **Reconciliation**: recorded when you reconcile transactions against a statement

### Transactions

Every transaction has a positive `amount` and an explicit `type` field (`income` / `expense` / `transfer`). Transactions represent actual money movements that have happened.

### Recurring Rules

Templates that define expected future transactions. These are **never materialized as rows** — projections are computed at query time by evaluating active rules against a date range.

## Balance Calculations

### Current Balance

Start from the most recent snapshot, then apply all transactions since:

```
current_balance(account) =
    latest_snapshot.balance
  + sum(transactions WHERE type = 'income'
        AND account = account
        AND date > latest_snapshot.date)
  + sum(transactions WHERE type = 'refund'
        AND account = account
        AND date > latest_snapshot.date)
  - sum(transactions WHERE type = 'expense'
        AND account = account
        AND date > latest_snapshot.date)
  ± transfers (in/out)
```

If no snapshot exists, you need either an initial snapshot or a complete transaction history from account opening.

### Projected Balance

To project balance on a future date, start from the current balance and apply recurring rules forward:

```
projected_balance(account, target_date) =
    current_balance(account)
  + sum(recurring_rule_occurrences(today, target_date)
        WHERE account = account)
```

Recurring rule occurrences are generated virtually — the system iterates each active rule's schedule and produces synthetic income/expense entries without storing them.

### Historical Balance

To get the balance at any past date, walk backward from the nearest snapshot:

```
historical_balance(account, past_date) =
    nearest_snapshot_after(past_date).balance
  - sum(transactions BETWEEN past_date AND snapshot.date)
```

Or walk forward from the nearest snapshot before that date. Use whichever snapshot is closer.

## Reconciliation

When snapshots and transaction sums disagree, that's useful information:

```
expected = snapshot.balance + sum(transactions since snapshot)
actual   = new_snapshot.balance
drift    = actual - expected
```

- **drift = 0**: books are clean.
- **drift ≠ 0**: something was missed — an unlogged transaction, a bank fee, a rounding error.

The system should surface drift as a reconciliation prompt, not silently correct it. The user decides whether to:
1. Add the missing transaction(s)
2. Record an adjustment transaction to zero out the drift
3. Ignore it (the new snapshot becomes the new anchor regardless)

## Snapshot Frequency

No fixed schedule required. Practical patterns:
- **Weekly**: check banking app on Sunday, log balances
- **On statement**: when a bank statement arrives, snapshot + reconcile
- **On login**: if bank import is automated, snapshot on every sync
- **Ad hoc**: whenever you feel like it

More snapshots = smaller windows where drift can accumulate = easier reconciliation.

## Edge Cases

- **New account**: create account + initial snapshot in one step. No transactions needed yet.
- **No snapshots at all**: balance is purely transaction-derived. Requires complete history or an explicit starting balance (which is just a snapshot with `date = account_open_date`).
- **Multiple snapshots same day**: use the latest one (by creation time).
- **Transfers between accounts**: a single transfer transaction decreases one account and increases another. Both sides must be accounted for in balance calculations.
