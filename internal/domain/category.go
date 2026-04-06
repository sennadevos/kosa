package domain

type CategoryType string

const (
	CategoryIncome  CategoryType = "income"
	CategoryExpense CategoryType = "expense"
	CategoryNeutral CategoryType = "neutral"
)

type Category struct {
	ID   string
	Name string
	Type CategoryType
}

type Tag struct {
	ID   string
	Name string
}
