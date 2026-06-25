package repositories

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type LessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) *LessonRepository {
	return &LessonRepository{db: db}
}

func (r *LessonRepository) FindAllByModule(moduleID int) ([]models.ModuleLesson, error) {
	var links []models.ModuleLesson
	err := r.db.Where("module_id = ?", moduleID).
		Preload("Lesson.LessonExercises.Exercise").
		Order("order_index asc").
		Find(&links).Error
	return links, err
}

func (r *LessonRepository) FindByID(id int) (*models.Lesson, error) {
	var lesson models.Lesson
	err := r.db.Preload("LessonExercises.Exercise").First(&lesson, id).Error
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (r *LessonRepository) Create(lesson *models.Lesson) error {
	return r.db.Create(lesson).Error
}

func (r *LessonRepository) Update(lesson *models.Lesson) error {
	return r.db.Save(lesson).Error
}

func (r *LessonRepository) Delete(id int) error {
	return r.db.Delete(&models.Lesson{}, id).Error
}

func (r *LessonRepository) FindAll() ([]models.Lesson, error) {
	var lessons []models.Lesson
	err := r.db.Order("title asc").Find(&lessons).Error
	return lessons, err
}

func (r *LessonRepository) FindModulesByLesson(lessonID int) ([]models.ModuleLesson, error) {
	var links []models.ModuleLesson
	err := r.db.Where("lesson_id = ?", lessonID).
		Preload("Module").
		Order("order_index asc").
		Find(&links).Error
	return links, err
}

func (r *LessonRepository) FindSharedLessons(moduleID int) ([]models.SharedLessonInfo, error) {
	var localPivots []models.ModuleLesson
	if err := r.db.Where("module_id = ?", moduleID).
		Preload("Lesson").
		Order("order_index asc").
		Find(&localPivots).Error; err != nil {
		return nil, err
	}
	if len(localPivots) == 0 {
		return nil, nil
	}

	lessonIDs := make([]int, 0, len(localPivots))
	for _, p := range localPivots {
		lessonIDs = append(lessonIDs, p.LessonID)
	}

	var sharedIDs []int
	r.db.Model(&models.ModuleLesson{}).
		Select("lesson_id").
		Where("lesson_id IN ?", lessonIDs).
		Group("lesson_id").
		Having("COUNT(DISTINCT module_id) > 1").
		Pluck("lesson_id", &sharedIDs)

	if len(sharedIDs) == 0 {
		return nil, nil
	}

	var otherPivots []models.ModuleLesson
	r.db.Where("lesson_id IN ? AND module_id != ?", sharedIDs, moduleID).
		Preload("Module").
		Order("module_id").
		Find(&otherPivots)

	otherModulesMap := make(map[int][]models.ModuleRef)
	for _, op := range otherPivots {
		otherModulesMap[op.LessonID] = append(otherModulesMap[op.LessonID], models.ModuleRef{
			ID:    op.Module.ID,
			Title: op.Module.Title,
		})
	}

	titleMap := make(map[int]string, len(localPivots))
	for _, p := range localPivots {
		titleMap[p.LessonID] = p.Lesson.Title
	}

	results := make([]models.SharedLessonInfo, 0, len(sharedIDs))
	for _, sid := range sharedIDs {
		results = append(results, models.SharedLessonInfo{
			LessonID:     sid,
			LessonTitle:  titleMap[sid],
			OtherModules: otherModulesMap[sid],
		})
	}

	return results, nil
}

func (r *LessonRepository) FindExercisesByLesson(lessonID int) ([]models.LessonExercise, error) {
	var links []models.LessonExercise
	err := r.db.Where("lesson_id = ?", lessonID).
		Preload("Exercise").
		Order("order_index asc").
		Find(&links).Error
	return links, err
}
