# Loan Tracking

## Overview

Loans track money owed between you and other people or entities. Each loan has an explicit direction (`payable` = you owe them, `receivable` = they owe you), a counterparty referenced by name and optional external URI, and optional interest terms.

## Loan Schema

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

## Interest Models

### No Interest (`interest_type = none`)

The simplest case. You lent someone €100, they pay back €100. Most informal loans between friends.

```
total_owed = original_amount
remaining  = original_amount - sum(payments)
```

### Flat Interest (`interest_type = flat`)

Interest is calculated once on the full original amount, regardless of how long the loan is outstanding or how payments are structured. Common for informal loans with a simple "pay me back a bit extra" agreement.

```
interest   = original_amount × (interest_rate / 100)
total_owed = original_amount + interest
remaining  = total_owed - sum(payments)
```

**Example**: You lend €1,000 at 5% flat interest.
- Interest = €50
- Total owed = €1,050
- They pay €200, then €300, then €550 → settled

The interest does not change based on time or remaining balance. It's a fixed fee on top of the principal.

### Periodic Interest (`interest_type = periodic`)

Interest accrues on the **remaining balance** at the end of each period. This is how most formal loans work — credit cards, mortgages, personal loans from banks.

**No interest state is stored.** The remaining balance is always computed by replaying the loan timeline: start from the principal, apply interest at each period boundary, subtract payments where they fall. This is a pure function of the loan terms + payment dates/amounts.

```
replay(principal, rate, period, start_date, payments[]):
    balance = principal
    for each period boundary from start_date to today:
        balance += balance × (rate / 100)        # interest accrues
        for each payment within this period:
            balance -= payment.amount             # payments reduce balance
    return balance
```

This avoids storing derived state that can get out of sync. The database holds facts (loan terms, payment records), the system computes the rest.

**Example**: €1,000 at 2% monthly interest, paid in 3 monthly installments.

| Month | Opening Balance | Interest (2%) | Payment | Closing Balance |
|-------|-----------------|---------------|---------|-----------------|
| 1     | 1,000.00        | 20.00         | 350.00  | 670.00          |
| 2     | 670.00          | 13.40         | 350.00  | 333.40          |
| 3     | 333.40          | 6.67          | 340.07  | 0.00            |

Total paid: €1,040.07 (€40.07 in interest).

Interest decreases as the balance decreases — paying early saves money. The replay function produces this table on demand from just the loan record and its three payment rows.

## Payments

### LoanPayments Schema

| Field   | Type   | Required | Description                                        |
|---------|--------|----------|----------------------------------------------------|
| id      | auto   | auto     |                                                    |
| loan    | link   | yes      | Which loan this pays against                       |
| date    | date   | no       | Payment date (defaults to today)                   |
| amount  | number | yes      | Amount paid                                        |
| account | link   | no       | Which account the money moved through              |
| notes   | text   | no       |                                                    |

### Partial Payments

Every payment is partial by default — it reduces the remaining balance. Full settlement happens when the remaining balance hits zero (or close enough for rounding).

For no-interest and flat-interest loans:
```
remaining = total_owed - sum(payments)
```

For periodic-interest loans:
```
remaining = replay(principal, rate, period, start_date, payments)
```

In all cases, when `remaining <= 0`, the loan is settled.

### Payment + Transaction Link

Each loan payment should also generate a corresponding transaction so it appears in cash flow:

- **Payable loan payment** → `expense` transaction on the paying account
- **Receivable loan payment** → `income` transaction on the receiving account

The transaction references the loan (via a `loan` link field) so it's clear this isn't regular spending/income.

## Computed Fields

These are derived at query time, not stored:

| Field             | Calculation                                                    |
|-------------------|----------------------------------------------------------------|
| `total_interest`  | Flat: `original × rate%`. Periodic: derived from replay.       |
| `total_owed`      | No interest / flat: `original_amount + total_interest`. Periodic: `original_amount + total_interest` (from replay). |
| `total_paid`      | `sum(loan_payments.amount)`                                    |
| `remaining`       | No interest / flat: `total_owed - total_paid`. Periodic: output of `replay()`. |
| `is_overdue`      | `due_date < today AND NOT is_settled`                          |
| `next_payment_due`| Derived from payment schedule if one exists                    |

## Examples

### Informal loan to a friend

```
type:               receivable
counterparty_name:  Jan
description:        Bike repair
original_amount:    150.00
interest_type:      none
due_date:           2026-05-01
```

Jan pays you back €75 twice → settled.

### Lending with flat interest

```
type:               receivable
counterparty_name:  Lisa
description:        Moving costs
original_amount:    500.00
interest_type:      flat
interest_rate:      10
due_date:           2026-09-01
```

Total owed: €550. Lisa pays in whatever chunks she can.

### Periodic interest (formal loan)

```
type:               payable
counterparty_name:  ABN AMRO
counterparty_uri:   https://contacts.example.com/api/entities/abn-amro
description:        Personal loan
original_amount:    5000.00
interest_type:      periodic
interest_rate:      1.5
interest_period:    monthly
due_date:           2027-04-01
```

1.5% monthly on remaining balance. Each payment reduces what future interest accrues on.

## Edge Cases

- **Overpayment**: if `total_paid > total_owed`, the excess could be flagged or auto-refunded depending on direction. For now, just mark as settled and note the overpayment.
- **Forgiven loans**: set `is_settled = true` manually, add a note. No payment needed.
- **Renegotiated terms**: update `interest_rate`, `due_date`, or `original_amount` and add a note explaining the change. Previous payments still count.
- **Currency mismatch**: if the loan is in a different currency than your accounts, the payment amount in your account currency may differ from the loan reduction amount. Store both if needed (payment in account currency, loan reduction in loan currency).
