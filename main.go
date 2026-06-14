package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/ebitengine/hideconsole"
	"github.com/go-vgo/robotgo"
	"github.com/something-that-is-cool/zutil/app"
	"github.com/something-that-is-cool/zutil/internal/logger"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
)

const DisableDebugLogs = false

var config = app.Config{
	Process:          "Minecraft.Windows.exe",
	Logger:           createLogger(),
	SupportedVersion: win.ProcessVersion{Major: 1, Minor: 1, Patch: 5},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := config.Logger.With("src", "initial-logs")
	log.Info("creating app instance...")

	a, err := config.New(ctx)
	if err != nil {
		doPanic(fmt.Errorf("error creating app: %w", err))
	}
	log.Info("created app instance.")
	defer closeApp(log, a)

	log.Info("starting app...")
	if err = a.Run(); err != nil {
		doPanic(fmt.Errorf("error running app: %w", err))
	}
}

func closeApp(l *slog.Logger, a *app.App) {
	if err := a.Close(); err != nil && !errors.Is(err, e.ErrAlreadyClosed) {
		l.Error("close app", "err", err.Error())
	}
}

func doPanic(v any) {
	msg := misc.JoinNewLine(
		fmt.Sprint(v),
		"",
		fmt.Sprintf("Please make sure you're running %q with version %q", config.Process, config.SupportedVersion.String()),
	)
	robotgo.Alert("Program exited with error (panic)", msg)
	os.Exit(1)
}

//goland:noinspection GoBoolExpressions
func createLogger() *slog.Logger {
	return logger.NewPrettySlogger(os.Stdout, logger.Level(!DisableDebugLogs))
}
