package customer

import (
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

type CustomerService interface {
	ConnectNutritionApp() (*string, error)
	VerifyToken(tokenSecret string) error
}

type Service struct {
	nutritionDiaryAPI vendor.NutritionDiary
}

func NewService(nutritionDiaryAPI vendor.NutritionDiary) *Service {
	return &Service{
		nutritionDiaryAPI: nutritionDiaryAPI,
	}
}

func (s *Service) ConnectNutritionApp() (*string, error) {
	oauth, err := s.nutritionDiaryAPI.GetToken()

	if err != nil {
		return nil, eris.Wrap(err, "Error getting the unauthorized token")
	}

	authURL, err := s.nutritionDiaryAPI.AuthorizeToken(oauth)

	if err != nil {
		return nil, eris.Wrap(err, "Error authorizing the token")
	}

	return authURL, nil
}

func (s *Service) VerifyToken(tokenSecret string) error {
	return nil
}
