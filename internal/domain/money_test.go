package domain

import (
	"encoding/json"
	"testing"
)

func TestNewAmount(t *testing.T) {
	a, err := NewAmount("12.50")
	if err != nil {
		t.Fatal(err)
	}
	if a.Format() != "12.50" {
		t.Fatalf("expected 12.50, got %s", a.Format())
	}
}

func TestNewAmountInvalid(t *testing.T) {
	_, err := NewAmount("abc")
	if err == nil {
		t.Fatal("expected error for invalid amount")
	}
}

func TestAmountArithmetic(t *testing.T) {
	a, _ := NewAmount("100.00")
	b, _ := NewAmount("33.33")

	sum := a.Add(b)
	if sum.Format() != "133.33" {
		t.Fatalf("expected 133.33, got %s", sum.Format())
	}

	diff := a.Sub(b)
	if diff.Format() != "66.67" {
		t.Fatalf("expected 66.67, got %s", diff.Format())
	}
}

func TestAmountJSON(t *testing.T) {
	a, _ := NewAmount("42.10")
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"42.10"` {
		t.Fatalf("expected \"42.10\", got %s", string(data))
	}

	var b Amount
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatal(err)
	}
	if b.Format() != "42.10" {
		t.Fatalf("expected 42.10, got %s", b.Format())
	}
}

func TestAmountJSONFromFloat(t *testing.T) {
	var a Amount
	if err := json.Unmarshal([]byte("12.5"), &a); err != nil {
		t.Fatal(err)
	}
	if a.Format() != "12.50" {
		t.Fatalf("expected 12.50, got %s", a.Format())
	}
}

func TestZeroAmount(t *testing.T) {
	z := ZeroAmount()
	if !z.IsZero() {
		t.Fatal("expected zero")
	}
	if z.Format() != "0.00" {
		t.Fatalf("expected 0.00, got %s", z.Format())
	}
}
