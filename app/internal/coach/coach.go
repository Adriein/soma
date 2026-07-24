package coach

import (
	"time"

	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/internal/meal"
)

type Assessment struct {
	ID      int
	Content string
	DateAdd time.Time
}

type DiaryEntry struct {
	Date  time.Time
	Meals []*meal.Meal
}

type AssessmentData struct {
	Days    int
	Profile *customer.Customer
	Diet    []*DiaryEntry
	Sources string
}
