package main

import (
	"context"
	_ "embed"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-vgo/robotgo"
	"github.com/k4ties/sensboost/app"
	"github.com/k4ties/sensboost/pkg/embeddable"
)

var (
	//go:embed config.yaml
	configYAML []byte
	// config ...
	config = embeddable.MustExtract[app.Config](embeddable.YAML, configYAML)
)

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
	msg := strings.Join([]string{
		fmt.Sprint(v),
		"",
		"Make sure you're running Minecraft Pocket Edition with version 1.1.5",
	}, "\n")
	robotgo.Alert("Program exited with error (panic)", msg, "OK")
	panic(v)
}
