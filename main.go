package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-vgo/robotgo"
	"github.com/something-that-is-cool/zutil/app"
	"github.com/something-that-is-cool/zutil/internal/misc"
)

var config = app.Config{
	Logger:  slog.Default(),
	Process: "Minecraft.Windows.exe",
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.SetLogLoggerLevel(slog.LevelDebug)

	log := config.Logger.With("src", "initial-logs")
	log.Info("creating app instance...")

	a, err := config.New(ctx)
	if err != nil {
		doPanic(fmt.Errorf("error creating app: %w", err))
	}
	log.Info("created app instance.")
	defer a.Close(true) //nolint:errcheck

	log.Info("starting app...")
	if err = a.Run(); err != nil {
		doPanic(fmt.Errorf("error running app: %w", err))
	}
}

func doPanic(v any) {
	msg := misc.JoinNewLine(
		fmt.Sprint(v),
		"",
		"Please make sure you're running Minecraft Pocket Edition with version 1.1.5",
	)
	robotgo.Alert("Program exited with error (panic)", msg)
	os.Exit(1)
}
