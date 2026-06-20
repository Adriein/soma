package meal

type MealService interface {
	Get() ([]*Meal, error)
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Get() ([]*Meal, error) {
	return nil, nil
}
