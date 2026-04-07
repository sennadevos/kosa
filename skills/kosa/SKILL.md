---
name: kosa
description: Personal finance CLI. Use when the user wants to track spending, income, loans, balances, or manage their finances. TRIGGER when user mentions money, spending, expenses, balance, loans, splitting costs, transactions, kosa, or financial tracking.
allowed-tools: Bash(kosa *)
argument-hint: [natural language finance request]
---

# kosa — Personal Finance CLI

kosa tracks transactions, loans, balances, and recurring rules backed by Teable (PostgreSQL).

## Quick Reference

### Record transactions
```bash
kosa spend <amount> "<description>" --cat <category>
kosa s <amount> "<description>"                         # alias
kosa income <amount> "<description>" --cat <category>
kosa i <amount> "<description>"                         # alias
kosa transfer <amount> --from "<account>" --to "<account>"
kosa refund <amount> "<description>" --cat <category> --of <txn-id>
```

Optional flags for all transaction commands:
- `--account "<name>"` — override default account
- `--tag <name>` — repeatable
- `--date YYYY-MM-DD` — defaults to today
- `--foreign "<amount> <currency>"` — e.g. `--foreign "50 GBP"`
- `--ref "<reference>"` — bank reference / invoice ID
- `--notes "<text>"`

### Loans
```bash
kosa owe <amount> "<description>" --to "<name>"         # you owe someone
kosa lent <amount> "<description>" --to "<name>"        # someone owes you
kosa loan new "<desc>" <amount> --from/--to "<name>" [--interest periodic --rate 1.5 --period monthly --due YYYY-MM-DD]
kosa loan pay <loan-id> <amount>
kosa loan show <loan-id>
kosa loans [--payable | --receivable]
```

### Split costs
```bash
kosa split <total> "<description>" --cat <cat> --with "Name1,Name2,Name3"
kosa split <total> "<desc>" --with "Name1:150,Name2:120" --mine 130
```

### Balance & snapshots
```bash
kosa balance                                            # all accounts
kosa balance --account "<name>"                         # specific account
kosa balance --on YYYY-MM-DD                            # projected
kosa snapshot <balance> --account "<name>"
```

### Recurring rules
```bash
kosa recurring add "<name>" <amount> --type expense --freq monthly --day 15 --cat <cat>
kosa recurring list
kosa recurring pause <rule-id>
kosa recurring resume <rule-id>
```

### Query & analysis
```bash
kosa list [--limit N] [--cat <cat>] [--tag <tag>] [--type expense]
kosa search "<query>"
kosa summary --month [YYYY-MM]
kosa summary --from YYYY-MM-DD --to YYYY-MM-DD
kosa reconcile --month YYYY-MM
```

### Output formats
All commands support `--json` and `--toon` (token-optimized for LLMs).

## Behavior

- Run kosa commands directly — it's installed at `~/.local/bin/kosa`.
- When the user describes a transaction in natural language, map it to the right kosa command.
- Use `--toon` when you need to process output programmatically.
- For multiple operations, run kosa commands in parallel when they're independent.
- Present results concisely — summarize, don't dump raw output unless asked.
- Amounts are always positive. The command name determines direction (spend = expense, income = income).
- Loan-linked transactions are excluded from spending summaries.

## Request: $ARGUMENTS
