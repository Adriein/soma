package customer

import "github.com/adriein/soma/app/pkg/vendor"

type CustomerService interface {
	ConnectNutritionApp() (*string, error)
}

type Service struct {
	diary vendor.NutritionDiary
}

func NewService(diary vendor.NutritionDiary) *Service {
	return &Service{
		diary: diary,
	}
}

func (s *Service) ConnectNutritionApp() (*string, error) {
	oauth, err := s.diary.GetToken()

	if err != nil {
		return nil, err
	}

	authURL, err := s.diary.Authorize(oauth)

	if err != nil {
		return nil, err
	}

	return authURL, nil
}
