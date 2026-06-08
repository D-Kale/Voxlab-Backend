package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/voxlab/voxlab-backend/internal/config"
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

const (
	LiveSessionsPrefix = "live_sessions"
	UserStreakPrefix   = "user:streak:active"
	ErrorLogsKey       = "logs:errors"
)

var Ctx = context.Background()
var Redis *redis.Client

func ConnectRedis(cfg *config.Config) error {
	Redis = redis.NewClient(&redis.Options{
		Addr: cfg.Redis.URL,
	})

	if err := Redis.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("connecting to Redis: %w", err)
	}

	log.Println("Redis connection established")
	return nil
}

func GetRedis() *redis.Client {
	return Redis
}

type LiveSessionData struct {
	RoomID       string    `json:"room_id"`
	HostID       string    `json:"host_id"`
	Participants []string `json:"participants"`
	StartTime    time.Time `json:"start_time"`
	Status       string    `json:"status"`
}

func SetLiveSession(ctx context.Context, roomID string, data *LiveSessionData) error {
	key := fmt.Sprintf("%s:%s", LiveSessionsPrefix, roomID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("serializing session: %w", err)
	}

	err = Redis.HSet(ctx, key, map[string]interface{}{
		"data": string(jsonData),
	}).Err()
	if err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	err = Redis.Expire(ctx, key, 2*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("setting TTL: %w", err)
	}

	return nil
}

func GetLiveSession(ctx context.Context, roomID string) (*LiveSessionData, error) {
	key := fmt.Sprintf("%s:%s", LiveSessionsPrefix, roomID)

	result, err := Redis.HGet(ctx, key, "data").Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	var data LiveSessionData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("deserializing session: %w", err)
	}

	return &data, nil
}

func DeleteLiveSession(ctx context.Context, roomID string) error {
	key := fmt.Sprintf("%s:%s", LiveSessionsPrefix, roomID)
	return Redis.Del(ctx, key).Err()
}

func SetUserStreak(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s:%s", UserStreakPrefix, userID)

	err := Redis.Set(ctx, key, time.Now().Unix(), 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("saving streak: %w", err)
	}

	return nil
}

func GetUserStreak(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("%s:%s", UserStreakPrefix, userID)

	exists, err := Redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("checking streak: %w", err)
	}

	return exists == 1, nil
}

func TrackUserProgress(ctx context.Context, userID string, xp int, streakDays int) error {
	key := fmt.Sprintf("user:progress:%s", userID)

	data := map[string]interface{}{
		"xp":          xp,
		"streak_days": streakDays,
		"updated_at":  time.Now().Unix(),
	}

	return Redis.HMSet(ctx, key, data).Err()
}

func GetUserProgressCache(ctx context.Context, userID string) (map[string]string, error) {
	key := fmt.Sprintf("user:progress:%s", userID)

	return Redis.HGetAll(ctx, key).Result()
}

func IncrementUserXP(ctx context.Context, userID uuid.UUID, amount int) error {
	err := DB.Model(&models.User{}).Where("id = ?", userID).Update("xp", gorm.Expr("xp + ?", amount)).Error
	if err != nil {
		return fmt.Errorf("updating XP: %w", err)
	}

	key := fmt.Sprintf("user:progress:%s", userID.String())
	err = Redis.HIncrBy(ctx, key, "xp", int64(amount)).Err()
	if err != nil {
		return fmt.Errorf("updating XP cache: %w", err)
	}

	return nil
}

func LogError(ctx context.Context, errorMsg string) error {
	err := Redis.LPush(ctx, ErrorLogsKey, fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), errorMsg)).Err()
	if err != nil {
		return fmt.Errorf("logging error: %w", err)
	}

	err = Redis.LTrim(ctx, ErrorLogsKey, 0, 999).Err()
	if err != nil {
		return fmt.Errorf("truncating logs: %w", err)
	}

	return nil
}

func GetErrorLogs(ctx context.Context, count int64) ([]string, error) {
	result, err := Redis.LRange(ctx, ErrorLogsKey, 0, count-1).Result()
	if err != nil {
		return nil, fmt.Errorf("fetching logs: %w", err)
	}

	return result, nil
}
