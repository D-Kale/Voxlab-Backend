package services

import (
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type ExerciseService struct {
	repo               *repositories.ExerciseRepository
	lessonExerciseRepo *repositories.LessonExerciseRepository
}

func NewExerciseService(repo *repositories.ExerciseRepository, lessonExerciseRepo *repositories.LessonExerciseRepository) *ExerciseService {
	return &ExerciseService{repo: repo, lessonExerciseRepo: lessonExerciseRepo}
}

func (s *ExerciseService) GetAll() ([]models.Exercise, error) {
	return s.repo.FindAll()
}

func (s *ExerciseService) GetByID(id uuid.UUID) (*models.Exercise, error) {
	return s.repo.FindByID(id)
}

func (s *ExerciseService) Create(exercise *models.Exercise) error {
	return s.repo.Create(exercise)
}

func (s *ExerciseService) Update(exercise *models.Exercise) error {
	return s.repo.Update(exercise)
}

func (s *ExerciseService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

// LinkExerciseToLesson links an existing exercise to a lesson via the pivot table.
func (s *ExerciseService) LinkExerciseToLesson(lessonID int, exerciseID uuid.UUID) error {
	next, err := s.lessonExerciseRepo.GetNextOrderIndex(lessonID)
	if err != nil {
		return err
	}
	return s.lessonExerciseRepo.Link(lessonID, exerciseID, next)
}

// UnlinkExerciseFromLesson removes the link between an exercise and a lesson.
func (s *ExerciseService) UnlinkExerciseFromLesson(lessonID int, exerciseID uuid.UUID) error {
	return s.lessonExerciseRepo.Unlink(lessonID, exerciseID)
}

// ReorderExerciseInLesson updates the order_index of an exercise within a lesson.
func (s *ExerciseService) ReorderExerciseInLesson(lessonID int, exerciseID uuid.UUID, newOrder int) error {
	return s.lessonExerciseRepo.UpdateOrder(lessonID, exerciseID, newOrder)
}

// BatchReorderExercises updates order_index for multiple exercises in a lesson (batch).
func (s *ExerciseService) BatchReorderExercises(lessonID int, items []models.LessonExercise) error {
	return s.lessonExerciseRepo.BatchReorder(lessonID, items)
}

// GetLessonsByExercise returns all LessonExercise records for a given exercise.
func (s *ExerciseService) GetLessonsByExercise(exerciseID uuid.UUID) ([]models.LessonExercise, error) {
	return s.lessonExerciseRepo.FindByExercise(exerciseID)
}
