package main

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"os/signal"
	"payment-service/bin/pkg/log"
	"payment-service/bin/pkg/redis"
	"payment-service/bin/pkg/utils"
	"time"

	"payment-service/bin/config"

	paymentHandler "payment-service/bin/modules/billing/handlers"
	paymentRepoCommands "payment-service/bin/modules/billing/repositories/commands"
	paymentRepoQueries "payment-service/bin/modules/billing/repositories/queries"
	paymentUsecase "payment-service/bin/modules/billing/usecases"
	kafkaConfluent "payment-service/bin/pkg/kafka/confluent"

	"payment-service/bin/pkg/apm"
	"payment-service/bin/pkg/databases/mongodb"

	"payment-service/bin/pkg/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.elastic.co/apm/module/apmechov4"
)

func main() {
	apm.InitConnection()
	redis.LoadConfig()
	redis.InitConnection()
	mongodb.InitConnection()
	kafkaConfluent.InitKafkaConfig()
	log.Init()
	e := echo.New()
	e.Validator = &validator.CustomValidator{Validator: validator.New()}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper:          middleware.DefaultSkipper,
		Format:           `[ROUTE] ${time_rfc3339} | ${status} | ${latency_human} ${latency} | ${method} | ${uri}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}))
	e.Use(middleware.Recover())
	e.Use(apmechov4.Middleware(apmechov4.WithTracer(apm.GetTracer())))

	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	setConfluentEvents()

	setHttp(e)

	listenerPort := fmt.Sprintf(":%s", config.GetConfig().AppPort)
	e.Logger.Fatal(e.Start(listenerPort))

	server := &http.Server{
		Addr:         listenerPort,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.GetLogger().Info("main", "Server message-service is shutting down...", "gracefull", "")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.GetLogger().Info("main", fmt.Sprintf("Could not gracefully shutdown the server order-service: %v\n", err), "gracefull", "")
		}
		close(done)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.GetLogger().Info("main", fmt.Sprintf("Could not listen on %s: %v\n", config.GetConfig().AppPort, err), "gracefull", "")
	}

	<-done
	log.GetLogger().Info("main", fmt.Sprintf("Server %s stopped", config.GetConfig().AppName), "gracefull", "")
}

func setConfluentEvents() {
	redisClient := redis.GetClient()
	kafkaProducer, err := kafkaConfluent.NewProducer(kafkaConfluent.GetConfig().GetKafkaConfig(), log.GetLogger())
	if err != nil {
		panic(err)
	}
	paymentQueryMongoRepo := paymentRepoQueries.NewQueryMongodbRepository(mongodb.NewMongoDBLogger(mongodb.GetSlaveConn(), mongodb.GetSlaveDBName(), log.GetLogger()))
	paymentCommandRepo := paymentRepoCommands.NewCommandMongodbRepository(mongodb.NewMongoDBLogger(mongodb.GetSlaveConn(), mongodb.GetSlaveDBName(), log.GetLogger()))
	paymentCommandUsecase := paymentUsecase.NewCommandUsecase(paymentQueryMongoRepo, paymentCommandRepo, redisClient, kafkaProducer)
	paymentConsumer, err := kafkaConfluent.NewConsumer(kafkaConfluent.GetConfig().GetKafkaConfig(), log.GetLogger())

	paymentHandler.InitPaymentEventHandler(paymentCommandUsecase, paymentConsumer)

	if err != nil {
		log.GetLogger().Error("main", "error registerNewConsumer", "setConfluentEvents", err.Error())
	}
}

func setHttp(e *echo.Echo) {
	e.GET("/v1/health-check", func(c echo.Context) error {
		log.GetLogger().Info("main", "This service is running properly", "setConfluentEvents", "")
		return utils.Response(nil, "This service is running properly", 200, c)
	})
	redisClient := redis.GetClient()
	kafkaProducer, err := kafkaConfluent.NewProducer(kafkaConfluent.GetConfig().GetKafkaConfig(), log.GetLogger())
	if err != nil {
		panic(err)
	}
	paymentQueryMongoRepo := paymentRepoQueries.NewQueryMongodbRepository(mongodb.NewMongoDBLogger(mongodb.GetSlaveConn(), mongodb.GetSlaveDBName(), log.GetLogger()))
	paymentCommandRepo := paymentRepoCommands.NewCommandMongodbRepository(mongodb.NewMongoDBLogger(mongodb.GetSlaveConn(), mongodb.GetSlaveDBName(), log.GetLogger()))

	paymentQueryUsecase := paymentUsecase.NewQueryUsecase(paymentQueryMongoRepo, redisClient)
	paymentCommandUsecase := paymentUsecase.NewCommandUsecase(paymentQueryMongoRepo, paymentCommandRepo, redisClient, kafkaProducer)

	paymentHandler.InitbillingHttpHandler(e, paymentQueryUsecase, paymentCommandUsecase)
}
