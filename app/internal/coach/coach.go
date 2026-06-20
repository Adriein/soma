package coach

import (
	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/internal/meal"
)

type ActionPlan struct {
	ID      int
	Content string
}

type AssessmentData struct {
	Profile *customer.Customer
	Diet    []*meal.Meal
}
