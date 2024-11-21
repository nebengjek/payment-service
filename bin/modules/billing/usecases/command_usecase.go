package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	driver "payment-service/bin/modules/billing"
	walletModels "payment-service/bin/modules/wallet/models"

	"payment-service/bin/modules/billing/models"

	httpError "payment-service/bin/pkg/http-error"
	kafkaPkgConfluent "payment-service/bin/pkg/kafka/confluent"
	"payment-service/bin/pkg/log"
	"payment-service/bin/pkg/utils"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	key := fmt.Sprintf("USER:ROUTE:%s", tripOrder.PassengerID)
	var tripPlan models.RouteSummary
	redisData, redisErr := c.redisClient.Get(ctx, key).Result()
	if redisErr != nil {
		log.GetLogger().Error("command_usecase", fmt.Sprintf("Redis Get Error: %v", redisErr), "GetRouteSummary", utils.ConvertString(redisErr))
		errObj := httpError.InternalServerError(fmt.Sprintf("Error getting data from Redis: %v", redisErr))
		return errObj
	}

	err := json.Unmarshal([]byte(redisData), &tripPlan)
	if err != nil {
		log.GetLogger().Error("command_usecase", err.Error(), "Unmarshal RouteSummary", utils.ConvertString(err))
		errObj := httpError.InternalServerError(fmt.Sprintf("Error unmarshalling tripData: %v", err))
		return errObj
	}

	totalFare, adminFee, driverEarnings := CalculateFinalFare(payload.RealDistance*3000, payload.FarePercentage)
	if totalFare > tripPlan.MaxPrice {
		totalFare = tripPlan.MaxPrice
		adminFee = totalFare * 0.05
		driverEarnings = totalFare - adminFee
	}
	transaction := models.Transaction{
		OrderID:        tripOrder.OrderID,
		PassengerID:    tripOrder.PassengerID,
		DriverID:       tripOrder.DriverID,
		TotalFare:      totalFare,
		AdminFee:       adminFee,
		DriverEarnings: driverEarnings,
		PaymentMethod:  "LinkAja",
		Status:         tripOrder.Status,
		Timestamp:      time.Now(),
	}

	// ambil saldo wallet passanger
	walletPsg := <-c.driverRepositoryQuery.Findwallet(ctx, transaction.PassengerID)
	if walletPsg.Error != nil {
		errObj := httpError.NotFound("wallet not found")
		log.GetLogger().Error("command_usecase", "wallet not found", "CompletedTrip", utils.ConvertString(walletPsg.Error))
		return errObj
	}
	walletPassanger := walletPsg.Data.(walletModels.Wallet)
	walletPassanger.Balance -= totalFare
	walletPassanger.TransactionLog = append(walletPassanger.TransactionLog, walletModels.TransactionLog{
		TransactionID: primitive.NewObjectID().Hex(),
		Amount:        totalFare,
		Type:          "pay nebengjek",
		Description:   "debit nebengjek via API",
		Timestamp:     time.Now(),
	})
	walletPassanger.LastUpdated = time.Now()
	updateWallet := <-c.driverRepositoryCommand.UpdateWallet(ctx, walletPassanger)
	if updateWallet.Error != nil {
		errObj := httpError.InternalServerError(fmt.Sprintf("Error: %v, Please try again later", updateWallet.Error))
		log.GetLogger().Error("command_usecase", fmt.Sprintf("Error: %v, Please try again later", updateWallet.Error), "CompletedTrip", utils.ConvertString(updateWallet.Error))
		return errObj
	}

	// add to wallet driver
	driver := <-c.driverRepositoryQuery.FindDriver(ctx, tripOrder.DriverID)
	if driver.Error != nil {
		errObj := httpError.NotFound("wallet not found")
		log.GetLogger().Error("command_usecase", "wallet not found", "CompletedTrip", utils.ConvertString(driver.Error))
		return errObj
	}
	infoDriver := driver.Data.(models.User)
	walletDriver := <-c.driverRepositoryQuery.Findwallet(ctx, infoDriver.UserID)
	if walletPsg.Error != nil {
		errObj := httpError.NotFound("wallet not found")
		log.GetLogger().Error("command_usecase", "wallet not found", "CompletedTrip", utils.ConvertString(walletDriver.Error))
		return errObj
	}
	wallet := walletDriver.Data.(walletModels.Wallet)
	wallet.Balance += driverEarnings
	wallet.TransactionLog = append(wallet.TransactionLog, walletModels.TransactionLog{
		TransactionID: primitive.NewObjectID().Hex(),
		Amount:        totalFare,
		Type:          "credit nebengjek",
		Description:   "credit nebengjek via API",
		Timestamp:     time.Now(),
	})
	wallet.LastUpdated = time.Now()
	updateWalletDriver := <-c.driverRepositoryCommand.UpdateWallet(ctx, wallet)
	if updateWalletDriver.Error != nil {
		errObj := httpError.InternalServerError(fmt.Sprintf("Error: %v, Please try again later", updateWalletDriver.Error))
		log.GetLogger().Error("command_usecase", fmt.Sprintf("Error: %v, Please try again later", updateWalletDriver.Error), "CompletedTrip", utils.ConvertString(updateWalletDriver.Error))
		return errObj
	}
	// insert adminfee
	adminFeeRecord := models.AdminFee{
		OrderID:       payload.OrderID,
		PassengerID:   tripOrder.PassengerID,
		DriverID:      tripOrder.DriverID,
		TripAmount:    totalFare,
		AdminFee:      adminFee,
		CollectedAt:   time.Now(),
		PaymentMethod: "LinkAja",
	}
	insertEarning := <-c.driverRepositoryCommand.InsertEarnings(ctx, adminFeeRecord)
	if insertEarning.Error != nil {
		errObj := httpError.InternalServerError("insert earning failed")
		return errObj
	}

	createBill := <-c.driverRepositoryCommand.InsertBilling(ctx, transaction)
	if createBill.Error != nil {
		errObj := httpError.InternalServerError("create billing failed")
		return errObj
	}
	c.redisClient.Del(ctx, key)
	return nil
}

func CalculateFinalFare(baseFare, discountPercentage float64) (totalFare, adminFee, driverEarnings float64) {
	totalFare = baseFare * (discountPercentage / 100)
	adminFee = totalFare * 0.05
	driverEarnings = totalFare - adminFee
	return
}
