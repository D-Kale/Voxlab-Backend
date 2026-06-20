package services

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type LessonService struct {
	repo *repositories.LessonRepository
}

func NewLessonService(repo *repositories.LessonRepository) *LessonService {
	return &LessonService{repo: repo}
}

func (s *LessonService) GetAllByModule(moduleID int) ([]models.ModuleLesson, error) {
	return s.repo.FindAllByModule(moduleID)
}

func (s *LessonService) GetByID(id int) (*models.Lesson, error) {
	return s.repo.FindByID(id)
}

func (s *LessonService) Create(lesson *models.Lesson) error {
	return s.repo.Create(lesson)
}

func (s *LessonService) Update(lesson *models.Lesson) error {
	return s.repo.Update(lesson)
}

func (s *LessonService) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *LessonService) GetAll() ([]models.Lesson, error) {
	return s.repo.FindAll()
}

func (s *LessonService) GetModulesByLesson(lessonID int) ([]models.ModuleLesson, error) {
	return s.repo.FindModulesByLesson(lessonID)
}

func (s *LessonService) GetExercisesByLesson(lessonID int) ([]models.LessonExercise, error) {
	return s.repo.FindExercisesByLesson(lessonID)
}
