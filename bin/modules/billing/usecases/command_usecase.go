package usecases

import (
	"context"
	"time"

	driver "payment-service/bin/modules/billing"
	"payment-service/bin/modules/billing/models"
	httpError "payment-service/bin/pkg/http-error"
	kafkaPkgConfluent "payment-service/bin/pkg/kafka/confluent"
	"payment-service/bin/pkg/log"
	"payment-service/bin/pkg/utils"

	"github.com/redis/go-redis/v9"
)

type commandUsecase struct {
	driverRepositoryQuery   driver.MongodbRepositoryQuery
	driverRepositoryCommand driver.MongodbRepositoryCommand
	redisClient             redis.UniversalClient
	kafkaProducer           kafkaPkgConfluent.Producer
}

func NewCommandUsecase(mq driver.MongodbRepositoryQuery, mc driver.MongodbRepositoryCommand, rc redis.UniversalClient, kp kafkaPkgConfluent.Producer) driver.UsecaseCommand {
	return &commandUsecase{
		driverRepositoryQuery:   mq,
		driverRepositoryCommand: mc,
		redisClient:             rc,
		kafkaProducer:           kp,
	}
}

func (c *commandUsecase) CreateBilling(ctx context.Context, payload models.TripOrderCompleted) error {
	trip := <-c.driverRepositoryQuery.FindActiveOrderPassanger(ctx, payload.OrderID)
	if trip.Error != nil {
		errObj := httpError.NotFound("The order has Notfound")
		log.GetLogger().Error("command_usecase", "The order has Notfound", "CompletedTrip", utils.ConvertString(trip.Error))
		return errObj
	}
	tripOrder := trip.Data.(models.TripOrder)
	totalFare, adminFee, driverEarnings := CalculateFinalFare(payload.RealDistance*3000, payload.FarePercentage)

	transaction := models.Transaction{
		OrderID:        tripOrder.OrderID,
		PassengerID:    tripOrder.PassengerID,
		DriverID:       tripOrder.DriverID,
		TotalFare:      totalFare,
		AdminFee:       adminFee,
		DriverEarnings: driverEarnings,
		PaymentMethod:  "LinkAja",
		Status:         "waiting-approval",
		Timestamp:      time.Now(),
	}

	sendNotif := <-c.driverRepositoryCommand.InsertBilling(ctx, transaction)
	if sendNotif.Error != nil {
		errObj := httpError.InternalServerError("create billing failed")
		return errObj
	}
	return nil
}

func CalculateFinalFare(baseFare, discountPercentage float64) (totalFare, adminFee, driverEarnings float64) {
	totalFare = baseFare * (discountPercentage / 100)
	adminFee = totalFare * 0.05
	driverEarnings = totalFare - adminFee
	return
}
