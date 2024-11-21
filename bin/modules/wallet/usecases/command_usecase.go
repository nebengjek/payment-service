package usecases

import (
	"context"
	"fmt"
	"time"

	driver "payment-service/bin/modules/wallet"
	"payment-service/bin/modules/wallet/models"
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

func (c *commandUsecase) TopUpWallet(ctx context.Context, userId string, payload models.TopUpRequest) utils.Result {
	var result utils.Result
	walletUser := <-c.driverRepositoryQuery.Findwallet(ctx, userId)
	if walletUser.Error != nil {
		// create new wallet
		newWallet := models.Wallet{
			UserID:      userId,
			Balance:     payload.Amount,
			LastUpdated: time.Now(),
			TransactionLog: []models.TransactionLog{
				{
					TransactionID: primitive.NewObjectID().Hex(),
					Amount:        payload.Amount,
					Type:          "topup",
					Description:   "Initial top-up",
					Timestamp:     time.Now(),
				},
			},
		}
		createWallet := <-c.driverRepositoryCommand.Insertwallet(ctx, newWallet)
		if createWallet.Error != nil {
			errObj := httpError.NewInternalServerError()
			errObj.Message = fmt.Sprintf("Error: %v, Please try again later", createWallet.Error)
			log.GetLogger().Error("command_usecase", errObj.Message, "CompletedTrip", utils.ConvertString(createWallet.Error))
			result.Error = errObj
			return result
		}
		result.Data = newWallet
		return result
	}
	wallet := walletUser.Data.(models.Wallet)
	wallet.Balance += payload.Amount
	wallet.TransactionLog = append(wallet.TransactionLog, models.TransactionLog{
		TransactionID: primitive.NewObjectID().Hex(),
		Amount:        payload.Amount,
		Type:          "topup",
		Description:   "Top-up via API",
		Timestamp:     time.Now(),
	})
	wallet.LastUpdated = time.Now()
	updateWallet := <-c.driverRepositoryCommand.UpdateWallet(ctx, wallet)
	if updateWallet.Error != nil {
		errObj := httpError.NewInternalServerError()
		errObj.Message = fmt.Sprintf("Error: %v, Please try again later", updateWallet.Error)
		log.GetLogger().Error("command_usecase", errObj.Message, "CompletedTrip", utils.ConvertString(updateWallet.Error))
		result.Error = errObj
		return result
	}
	result.Data = wallet
	return result
}

func CalculateFinalFare(baseFare, discountPercentage float64) (totalFare, adminFee, driverEarnings float64) {
	totalFare = baseFare * (discountPercentage / 100)
	adminFee = totalFare * 0.05
	driverEarnings = totalFare - adminFee
	return
}
