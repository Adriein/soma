package coach

type CoachService interface {
	Assessment(data *AssessmentData) *ActionPlan
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Assessment(data *AssessmentData) (*ActionPlan, error) {
	return nil, nil
}
