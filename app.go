package main

import (
	"context"
	"fmt"
	"inspection-service/api"
	"inspection-service/cluster/analyzer"
	"inspection-service/cluster/brigade"
	"inspection-service/cluster/file"
	"inspection-service/cluster/subscriber"
	"inspection-service/cluster/task"
	"inspection-service/config"
	dbinspection "inspection-service/database/inspection"
	"inspection-service/service/inspection"
	"io/fs"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sunshineOfficial/golib/db"
	"github.com/sunshineOfficial/golib/gohttp"
	"github.com/sunshineOfficial/golib/gohttp/goserver"
	"github.com/sunshineOfficial/golib/gokafka"
	"github.com/sunshineOfficial/golib/golog"
)

const (
	serviceName = "inspection-service"
	dbTimeout   = 15 * time.Second
)

type App struct {
	/* main */
	mainCtx  context.Context
	log      golog.Logger
	settings config.Settings

	/* http */
	server goserver.Server

	/* db */
	postgres           *sqlx.DB
	kafka              gokafka.Kafka
	inspectionProducer gokafka.Producer
	taskConsumer       gokafka.Consumer

	/* services */
	inspectionService *inspection.Service
}

func NewApp(mainCtx context.Context, log golog.Logger, settings config.Settings) *App {
	return &App{
		mainCtx:  mainCtx,
		log:      log,
		settings: settings,
	}
}

func (a *App) InitDatabases(fs fs.FS, path string) (err error) {
	postgresCtx, cancelPostgresCtx := context.WithTimeout(a.mainCtx, dbTimeout)
	defer cancelPostgresCtx()

	a.postgres, err = db.NewPgx(postgresCtx, a.settings.Databases.Postgres)
	if err != nil {
		return fmt.Errorf("init postgres: %w", err)
	}

	err = db.Migrate(fs, a.log, a.postgres, path)
	if err != nil {
		return fmt.Errorf("migrate postgres: %w", err)
	}

	a.kafka = gokafka.NewKafka(a.settings.Databases.Kafka.Brokers)

	a.inspectionProducer = a.kafka.Producer(a.settings.Databases.Kafka.Topics.Inspections)

	a.taskConsumer, err = a.kafka.Consumer(a.log.WithTags("taskConsumer"), func() (context.Context, context.CancelFunc) {
		return context.WithCancel(a.mainCtx)
	}, gokafka.WithTopic(a.settings.Databases.Kafka.Topics.Tasks), gokafka.WithConsumerGroup(serviceName))
	if err != nil {
		return fmt.Errorf("init task consumer: %w", err)
	}

	return nil
}

func (a *App) InitServices() error {
	inspectionRepository := dbinspection.NewRepository(a.postgres)

	inspectionPublisher := inspection.NewPublisher(a.mainCtx, a.inspectionProducer)

	httpClient := gohttp.NewClient(gohttp.WithTimeout(1 * time.Minute))

	analyzerClient := analyzer.NewClient(httpClient, a.settings.Cluster.AnalyzerService)
	subscriberClient := subscriber.NewClient(httpClient, a.settings.Cluster.SubscriberService)
	fileClient := file.NewClient(httpClient, a.settings.Cluster.FileService)
	taskClient := task.NewClient(httpClient, a.settings.Cluster.TaskService)
	brigadeClient := brigade.NewClient(httpClient, a.settings.Cluster.BrigadeService)

	a.inspectionService = inspection.NewService(
		inspectionRepository,
		inspectionPublisher,
		analyzerClient,
		subscriberClient,
		fileClient,
		taskClient,
		brigadeClient,
		a.settings.Templates,
	)

	return nil
}

func (a *App) InitServer() {
	sb := api.NewServerBuilder(a.mainCtx, a.log, a.settings)
	sb.AddDebug()
	sb.AddInspections(a.inspectionService)

	a.server = sb.Build()
}

func (a *App) Start() {
	a.server.Start()
	a.taskConsumer.Subscribe(a.inspectionService.SubscriberOnTaskEvent(a.mainCtx, a.log.WithTags("taskSubscriber")))
}

func (a *App) Stop(ctx context.Context) {
	consumerCtx, cancelConsumerCtx := context.WithTimeout(ctx, dbTimeout)
	defer cancelConsumerCtx()

	err := a.taskConsumer.Close(consumerCtx)
	if err != nil {
		a.log.Errorf("failed to close task consumer: %v", err)
	}

	a.server.Stop()

	producerCtx, cancelProducerCtx := context.WithTimeout(ctx, dbTimeout)
	defer cancelProducerCtx()

	err = a.inspectionProducer.Close(producerCtx)
	if err != nil {
		a.log.Errorf("failed to close inspection producer: %v", err)
	}

	err = a.postgres.Close()
	if err != nil {
		a.log.Errorf("failed to close postgres connection: %v", err)
	}
}
