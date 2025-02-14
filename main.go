package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"go-auth-service/src/infra/broker/nats"
	natsPub "go-auth-service/src/infra/broker/nats/publisher"
	"go-auth-service/src/infra/persistence/redis"
	redisServe "go-auth-service/src/infra/persistence/redis/service"
	mailWorker "go-auth-service/src/interface/broker/mail"

	usecase "go-auth-service/src/app/usecases"
	mailUC "go-auth-service/src/app/usecases/mail"
	userUC "go-auth-service/src/app/usecases/user"
	"go-auth-service/src/infra/config"
	ms_log "go-auth-service/src/infra/log"
	postgresDb "go-auth-service/src/infra/persistence/postgres"
	historyRepo "go-auth-service/src/infra/persistence/postgres/history"
	userRepo "go-auth-service/src/infra/persistence/postgres/user"
	"go-auth-service/src/interface/rest"
)

func main() {
	ctx := context.Background()

	conf := config.Make()

	isProd := false
	if conf.App.Environment == "PRODUCTION" {
		isProd = true
	}

	Nats := nats.NewNats()
	if !Nats.Status {
		defer Nats.Conn.Close()
	}
	natsPublisher := natsPub.NewPublisher(Nats)

	m := make(map[string]interface{})
	m["env"] = conf.App.Environment
	m["service"] = conf.App.Name
	logger := ms_log.NewLogInstance(
		ms_log.LogName(conf.Log.Name),
		ms_log.IsProduction(isProd),
		ms_log.LogAdditionalFields(m))

	postgresConnection, err := postgresDb.NewConnection(conf.SqlDb.Master, conf.SqlDb.Slave, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer func(l *logrus.Logger, conn *postgresDb.Connection) {
		err := conn.GetPrimaryMaster().Close()
		if err != nil {
			l.Errorf("error closing master database: %s", err)
		} else {
			l.Println("Master database closed successfully")
		}

		err = conn.GetPrimarySlave().Close()
		if err != nil {
			l.Errorf("error closing slave database: %s", err)
		} else {
			l.Println("Slave database closed successfully")
		}
	}(logger, postgresConnection)

	redisClient, err := redis.NewRedisClient(conf.Redis, logger)
	redisService := redisServe.NewServRedis(redisClient)

	userRepository := userRepo.NewUserRepository(postgresConnection)
	historyRepository := historyRepo.NewHistoryRepository(postgresConnection)

	// Inisialisasi use cases
	useCaseList := usecase.AllUseCases{
		UserUC: userUC.NewUserUseCase(natsPublisher, redisService, userRepository, historyRepository),
		MailUC: mailUC.NewMailUseCase(redisService, userRepository, historyRepository),
	}

	// * worker initialization *
	mailWorker.NewAuthWorker(Nats, useCaseList.MailUC)

	httpServer, err := rest.New(
		conf.Http,
		isProd,
		logger,
		useCaseList,
	)
	if err != nil {
		logger.Fatalf("Failed to initialize HTTP server: %v", err)
	}
	httpServer.Start(ctx)
}
