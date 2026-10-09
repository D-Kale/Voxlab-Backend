// Package logger centraliza el logging estructurado de la API.
//
// Reemplaza el uso disperso de `log` de stdlib para que cada línea tenga
// estructura (clave/valor) y un request_id que permita correlacionar una
// falla del servidor con la que vio el usuario en el navegador.
//
// Solo depende de log/slog (stdlib), no de librerías de terceros.
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/voxlab/voxlab-backend/internal/config"
)

var log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

// Init configura el logger global. Debe llamarse una vez, después de
// config.LoadConfig() y antes de levantar el router.
//
// En development emite texto legible; en production emite JSON para que un
// agregador de logs pueda parsear los campos.
func Init(cfg *config.Config) {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}

	var handler slog.Handler
	if strings.EqualFold(cfg.AppEnv, "production") {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
	log = slog.Default()

	log.Info("logger initialized",
		"level", cfg.LogLevel,
		"env", cfg.AppEnv,
		"format", handlerFormat(cfg.AppEnv),
	)
}

func handlerFormat(appEnv string) string {
	if strings.EqualFold(appEnv, "production") {
		return "json"
	}
	return "text"
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Debug(msg string, args ...any) { log.Debug(msg, args...) }
func Info(msg string, args ...any)  { log.Info(msg, args...) }
func Warn(msg string, args ...any)  { log.Warn(msg, args...) }
func Error(msg string, args ...any) { log.Error(msg, args...) }

// FromContext devuelve un context con el logger. Existe para que las capas
// inferiores (services, repositories) puedan loguear sin importar gin.
func FromContext(ctx context.Context) *slog.Logger {
	return slog.New(slog.Default().Handler())
}
