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
	"github.com/k4ties/sensboost/app"
	"github.com/k4ties/sensboost/internal/misc"
)

var config = app.Config{
	Logger:  slog.Default(),
	Process: "Minecraft.Windows.exe",
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a, err := config.New(ctx)
	if err != nil {
		doPanic(fmt.Errorf("error creating app: %w", err))
	}
	defer a.Close(true) //nolint:errcheck
	if err = a.Run(); err != nil {
		doPanic(fmt.Errorf("error running app: %w", err))
	}
}

func doPanic(v any) {
	msg := misc.JoinNewLine(
		fmt.Sprint(v),
		"-----",
		"Please make sure you're running Minecraft Pocket Edition with version 1.1.5",
	)
	robotgo.Alert("Program exited with error (panic)", msg, "OK", "окак")
	os.Exit(1)
}
