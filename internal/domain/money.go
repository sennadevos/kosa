package domain

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// Amount represents a monetary value. Always use this instead of float64.
type Amount struct {
	decimal.Decimal
}

func NewAmount(s string) (Amount, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Amount{}, fmt.Errorf("invalid amount %q: %w", s, err)
	}
	return Amount{d}, nil
}

func NewAmountFromFloat(f float64) Amount {
	return Amount{decimal.NewFromFloat(f)}
}

func NewAmountFromInt(i int64) Amount {
	return Amount{decimal.NewFromInt(i)}
}

func ZeroAmount() Amount {
	return Amount{decimal.Zero}
}

func (a Amount) Add(b Amount) Amount {
	return Amount{a.Decimal.Add(b.Decimal)}
}

func (a Amount) Sub(b Amount) Amount {
	return Amount{a.Decimal.Sub(b.Decimal)}
}

func (a Amount) Mul(b Amount) Amount {
	return Amount{a.Decimal.Mul(b.Decimal)}
}

func (a Amount) Div(b Amount) Amount {
	return Amount{a.Decimal.Div(b.Decimal)}
}

func (a Amount) IsPositive() bool {
	return a.Decimal.IsPositive()
}

func (a Amount) IsNegative() bool {
	return a.Decimal.IsNegative()
}

func (a Amount) IsZero() bool {
	return a.Decimal.IsZero()
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.StringFixed(2))
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var f float64
		if err := json.Unmarshal(data, &f); err != nil {
			return fmt.Errorf("cannot unmarshal amount: %s", string(data))
		}
		a.Decimal = decimal.NewFromFloat(f)
		return nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return err
	}
	a.Decimal = d
	return nil
}

func (a Amount) Format() string {
	return a.StringFixed(2)
}
