package domain

type AccountType string

const (
	AccountChecking   AccountType = "checking"
	AccountSavings    AccountType = "savings"
	AccountInvestment AccountType = "investment"
	AccountCreditCard AccountType = "credit_card"
	AccountCash       AccountType = "cash"
)

type Account struct {
	ID       string
	Name     string
	Type     AccountType
	Provider string
	Currency string
	IBAN     string
	Notes    string
}

type AccountInput struct {
	Name     string
	Type     AccountType
	Provider string
	Currency string
	IBAN     string
	Notes    string
}

type AccountFilter struct {
	Type *AccountType
}
