package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/app"
)

func stopAppOnOsSignal(a *app.App) {
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)

	<-osSignal

	a.Stop(context.Background())
}
