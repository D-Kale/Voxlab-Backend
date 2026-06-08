package educational

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetTracks() ([]Track, error) {
	return s.repo.FindAllTracks()
}
