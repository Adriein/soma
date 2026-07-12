package coach

import (
	"time"

	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/internal/meal"
)

type ActionPlan struct {
	ID      int
	Content string
}

type DiaryEntry struct {
	Date  time.Time
	Meals []*meal.Meal
}

type AssessmentData struct {
	Profile *customer.Customer
	Diet    []*DiaryEntry
	Sources string
}
