package meal

type MealService interface {
	Extract() []*Meal
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Extract() []*Meal {
	return nil
}
