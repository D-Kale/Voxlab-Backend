package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"strings"

	"github.com/chai2010/webp"
	"github.com/google/uuid"
	"golang.org/x/image/draw"

	"github.com/voxlab/voxlab-backend/internal/repositories"
	"github.com/voxlab/voxlab-backend/internal/storage"
)

type UploadService struct {
	store    *storage.Storage
	trackRepo   *repositories.TrackRepository
	moduleRepo  *repositories.ModuleRepository
	lessonRepo  *repositories.LessonRepository
	userRepo    *repositories.UserRepository
}

func NewUploadService(
	store *storage.Storage,
	trackRepo *repositories.TrackRepository,
	moduleRepo *repositories.ModuleRepository,
	lessonRepo *repositories.LessonRepository,
	userRepo *repositories.UserRepository,
) *UploadService {
	return &UploadService{
		store:      store,
		trackRepo:  trackRepo,
		moduleRepo: moduleRepo,
		lessonRepo: lessonRepo,
		userRepo:   userRepo,
	}
}

const (
	maxCoverWidth  = 1920
	maxCoverHeight = 1080
	maxAvatarSize  = 400
	webpQuality    = 80
)

type UploadResult struct {
	URL string `json:"url"`
}

func (s *UploadService) UploadTrackImage(ctx context.Context, trackID int, reader io.Reader) (*UploadResult, error) {
	track, err := s.trackRepo.FindByID(trackID)
	if err != nil {
		return nil, fmt.Errorf("track not found: %w", err)
	}

	img, err := decodeImage(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	img = resizeCover(img)

	path := fmt.Sprintf("tracks/%s.webp", uuid.New().String())
	url, err := s.processAndUpload(ctx, img, path)
	if err != nil {
		return nil, err
	}

	if track.IconURL != "" {
		_ = s.deleteOldFile(ctx, track.IconURL)
	}

	track.IconURL = url
	if err := s.trackRepo.Update(track); err != nil {
		return nil, fmt.Errorf("failed to update track: %w", err)
	}

	return &UploadResult{URL: url}, nil
}

func (s *UploadService) UploadModuleImage(ctx context.Context, moduleID int, reader io.Reader) (*UploadResult, error) {
	mod, err := s.moduleRepo.FindByID(moduleID)
	if err != nil {
		return nil, fmt.Errorf("module not found: %w", err)
	}

	img, err := decodeImage(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	img = resizeCover(img)

	path := fmt.Sprintf("modules/%s.webp", uuid.New().String())
	url, err := s.processAndUpload(ctx, img, path)
	if err != nil {
		return nil, err
	}

	if mod.ImageURL != "" {
		_ = s.deleteOldFile(ctx, mod.ImageURL)
	}

	mod.ImageURL = url
	if err := s.moduleRepo.Update(mod); err != nil {
		return nil, fmt.Errorf("failed to update module: %w", err)
	}

	return &UploadResult{URL: url}, nil
}

func (s *UploadService) UploadLessonImage(ctx context.Context, lessonID int, reader io.Reader) (*UploadResult, error) {
	lesson, err := s.lessonRepo.FindByID(lessonID)
	if err != nil {
		return nil, fmt.Errorf("lesson not found: %w", err)
	}

	img, err := decodeImage(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	img = resizeCover(img)

	path := fmt.Sprintf("lessons/%s.webp", uuid.New().String())
	url, err := s.processAndUpload(ctx, img, path)
	if err != nil {
		return nil, err
	}

	if lesson.ImageURL != "" {
		_ = s.deleteOldFile(ctx, lesson.ImageURL)
	}

	lesson.ImageURL = url
	if err := s.lessonRepo.Update(lesson); err != nil {
		return nil, fmt.Errorf("failed to update lesson: %w", err)
	}

	return &UploadResult{URL: url}, nil
}

func (s *UploadService) UploadAvatar(ctx context.Context, userID string, reader io.Reader) (*UploadResult, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	img, err := decodeImage(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	img = resizeAvatar(img)

	path := fmt.Sprintf("avatars/%s.webp", uuid.New().String())
	url, err := s.processAndUpload(ctx, img, path)
	if err != nil {
		return nil, err
	}

	if user.AvatarURL != "" {
		_ = s.deleteOldFile(ctx, user.AvatarURL)
	}

	user.AvatarURL = url
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &UploadResult{URL: url}, nil
}

func (s *UploadService) deleteOldFile(ctx context.Context, oldURL string) error {
	path := extractPathFromURL(oldURL)
	if path == "" {
		return nil
	}
	return s.store.Delete(ctx, path)
}

func (s *UploadService) processAndUpload(ctx context.Context, img image.Image, path string) (string, error) {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: webpQuality}); err != nil {
		return "", fmt.Errorf("webp encode: %w", err)
	}

	url, err := s.store.Upload(ctx, path, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", fmt.Errorf("minio upload: %w", err)
	}

	return url, nil
}

func decodeImage(r io.Reader) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func resizeCover(img image.Image) image.Image {
	return resizeToFit(img, maxCoverWidth, maxCoverHeight)
}

func resizeAvatar(img image.Image) image.Image {
	return resizeToFit(img, maxAvatarSize, maxAvatarSize)
}

func resizeToFit(img image.Image, maxW, maxH int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxW && h <= maxH {
		return img
	}

	rw := float64(maxW) / float64(w)
	rh := float64(maxH) / float64(h)
	ratio := rw
	if rh < rw {
		ratio = rh
	}
	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}

func extractPathFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	// Path format: /bucket/path/to/file → strip leading /bucket/
	parts := strings.SplitN(strings.TrimPrefix(parsed.Path, "/"), "/", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
