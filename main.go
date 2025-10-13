package main

import (
	"context"
	"inspection-service/config"
	"os"

	"github.com/shopspring/decimal"
	"github.com/sunshineOfficial/golib/golog"
	"github.com/sunshineOfficial/golib/goos"
)

func main() {
	configureDecimal()

	log := golog.NewLogger(serviceName)
	log.Debug("service up")

	settings, err := config.Get(log)
	if err != nil {
		log.Errorf("failed to get config: %v", err)
		return
	}

	mainCtx, cancelMainCtx := context.WithCancel(context.Background())
	defer cancelMainCtx()

	app := NewApp(mainCtx, log, settings)

	if err = app.InitDatabases(os.DirFS("./"), "database/migrations/postgres"); err != nil {
		log.Errorf("failed to init databases: %v", err)
		return
	}

	if err = app.InitServices(); err != nil {
		log.Errorf("failed to init services: %v", err)
		return
	}

	app.InitServer()

	app.Start()

	goos.WaitTerminate(mainCtx, app.Stop)
	log.Debug("service down")
}

func configureDecimal() {
	decimal.DivisionPrecision = 2
	decimal.MarshalJSONWithoutQuotes = true
}
