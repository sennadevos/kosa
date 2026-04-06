package app

import (
	"context"
	"testing"

	"github.com/sennadevos/kosa/internal/backend/mock"
	"github.com/sennadevos/kosa/internal/config"
	"github.com/sennadevos/kosa/internal/domain"
)

func testApp() (*App, *mock.Backend) {
	b := mock.New()
	b.SeedAccount(domain.Account{
		ID: "acc_1", Name: "ING Checking", Type: domain.AccountChecking, IsDefault: true,
	})
	b.SeedCategory(domain.Category{ID: "cat_1", Name: "groceries", Type: domain.CategoryExpense})
	b.SeedCategory(domain.Category{ID: "cat_2", Name: "salary", Type: domain.CategoryIncome})

	cfg := &config.Config{
		Defaults: config.DefaultsConfig{
			Account:  "ING Checking",
			Currency: "EUR",
		},
	}
	return New(b, cfg), b
}

func TestSpend(t *testing.T) {
	a, _ := testApp()
	ctx := context.Background()

	amount, _ := domain.NewAmount("4.50")
	txn, err := a.Spend(ctx, SpendInput{
		Amount:      amount,
		Description: "Coffee",
		Category:    "groceries",
	})
	if err != nil {
		t.Fatal(err)
	}
	if txn.Type != domain.TransactionExpense {
		t.Fatalf("expected expense, got %s", txn.Type)
	}
	if txn.Amount.Format() != "4.50" {
		t.Fatalf("expected 4.50, got %s", txn.Amount.Format())
	}
	if txn.Description != "Coffee" {
		t.Fatalf("expected Coffee, got %s", txn.Description)
	}
	if txn.CategoryID != "cat_1" {
		t.Fatalf("expected cat_1, got %s", txn.CategoryID)
	}
	if txn.AccountID != "acc_1" {
		t.Fatalf("expected acc_1, got %s", txn.AccountID)
	}
}

func TestSpendWithTags(t *testing.T) {
	a, _ := testApp()
	ctx := context.Background()

	amount, _ := domain.NewAmount("120.00")
	txn, err := a.Spend(ctx, SpendInput{
		Amount:      amount,
		Description: "Dinner",
		Tags:        []string{"date-night"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(txn.TagIDs) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(txn.TagIDs))
	}
}

func TestSpendWithForeign(t *testing.T) {
	a, _ := testApp()
	ctx := context.Background()

	amount, _ := domain.NewAmount("58.00")
	foreign, _ := domain.NewAmount("50.00")
	txn, err := a.Spend(ctx, SpendInput{
		Amount:          amount,
		Description:     "UK bookshop",
		ForeignAmount:   foreign,
		ForeignCurrency: "GBP",
	})
	if err != nil {
		t.Fatal(err)
	}
	if txn.ForeignCurrency != "GBP" {
		t.Fatalf("expected GBP, got %s", txn.ForeignCurrency)
	}
	if txn.ForeignAmount.Format() != "50.00" {
		t.Fatalf("expected 50.00, got %s", txn.ForeignAmount.Format())
	}
}

func TestIncome(t *testing.T) {
	a, _ := testApp()
	ctx := context.Background()

	amount, _ := domain.NewAmount("3200.00")
	txn, err := a.Income(ctx, IncomeInput{
		Amount:      amount,
		Description: "Salary",
		Category:    "salary",
	})
	if err != nil {
		t.Fatal(err)
	}
	if txn.Type != domain.TransactionIncome {
		t.Fatalf("expected income, got %s", txn.Type)
	}
}

func TestTransfer(t *testing.T) {
	a, b := testApp()
	b.SeedAccount(domain.Account{ID: "acc_2", Name: "ABN Savings", Type: domain.AccountSavings})
	ctx := context.Background()

	amount, _ := domain.NewAmount("500.00")
	txn, err := a.Transfer(ctx, TransferInput{
		Amount:      amount,
		FromAccount: "ING Checking",
		ToAccount:   "ABN Savings",
	})
	if err != nil {
		t.Fatal(err)
	}
	if txn.Type != domain.TransactionTransfer {
		t.Fatalf("expected transfer, got %s", txn.Type)
	}
	if txn.AccountID != "acc_1" {
		t.Fatalf("expected acc_1, got %s", txn.AccountID)
	}
	if txn.ToAccountID != "acc_2" {
		t.Fatalf("expected acc_2, got %s", txn.ToAccountID)
	}
}

func TestRefund(t *testing.T) {
	a, b := testApp()
	ctx := context.Background()

	// create original transaction
	amount, _ := domain.NewAmount("80.00")
	orig, _ := a.Spend(ctx, SpendInput{
		Amount:      amount,
		Description: "Zara jacket",
	})

	// refund it
	txn, err := a.Refund(ctx, RefundInput{
		Amount:      amount,
		Description: "Zara jacket return",
		RefundOfID:  orig.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if txn.Type != domain.TransactionRefund {
		t.Fatalf("expected refund, got %s", txn.Type)
	}
	if txn.RefundOfID != orig.ID {
		t.Fatalf("expected refund_of %s, got %s", orig.ID, txn.RefundOfID)
	}

	// verify both transactions exist
	all, _ := b.ListTransactions(ctx, domain.TransactionFilter{})
	if len(all) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(all))
	}
}

func TestList(t *testing.T) {
	a, _ := testApp()
	ctx := context.Background()

	amount, _ := domain.NewAmount("10.00")
	for i := 0; i < 5; i++ {
		a.Spend(ctx, SpendInput{Amount: amount, Description: "item"})
	}

	txns, err := a.ListTransactions(ctx, ListInput{Limit: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 3 {
		t.Fatalf("expected 3, got %d", len(txns))
	}
}

func TestSearch(t *testing.T) {
	a, _ := testApp()
	ctx := context.Background()

	amount, _ := domain.NewAmount("10.00")
	a.Spend(ctx, SpendInput{Amount: amount, Description: "Albert Heijn"})
	a.Spend(ctx, SpendInput{Amount: amount, Description: "Coffee"})
	a.Spend(ctx, SpendInput{Amount: amount, Description: "Albert Heijn Breda"})

	txns, err := a.SearchTransactions(ctx, "albert", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2, got %d", len(txns))
	}
}
