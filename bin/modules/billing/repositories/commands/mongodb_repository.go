package commands

import (
	"context"

	user "payment-service/bin/modules/billing"
	"payment-service/bin/modules/billing/models"
	walletModels "payment-service/bin/modules/wallet/models"
	"payment-service/bin/pkg/databases/mongodb"
	"payment-service/bin/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type commandMongodbRepository struct {
	mongoDb mongodb.MongoDBLogger
}

func NewCommandMongodbRepository(mongodb mongodb.MongoDBLogger) user.MongodbRepositoryCommand {
	return &commandMongodbRepository{
		mongoDb: mongodb,
	}
}

func (c commandMongodbRepository) NewObjectID(ctx context.Context) string {
	return primitive.NewObjectID().Hex()
}

func (c commandMongodbRepository) InsertBilling(ctx context.Context, data models.Transaction) <-chan utils.Result {
	output := make(chan utils.Result)

	go func() {
		defer close(output)

		err := c.mongoDb.UpsertOne(mongodb.UpsertOne{
			CollectionName: "billing",
			Filter: bson.M{
				"orderId":     data.OrderID,
				"passengerId": data.PassengerID,
			},
			Document: data,
		}, ctx)

		if err != nil {
			output <- utils.Result{
				Error: err,
			}
			return
		}

		output <- utils.Result{
			Data: data.ID,
		}
	}()

	return output
}

func (c commandMongodbRepository) InsertEarnings(ctx context.Context, data models.AdminFee) <-chan utils.Result {
	output := make(chan utils.Result)

	go func() {
		defer close(output)

		err := c.mongoDb.InsertOne(mongodb.InsertOne{
			CollectionName: "admin-earnings",
			Document:       data,
		}, ctx)

		if err != nil {
			output <- utils.Result{
				Error: err,
			}
			return
		}

		output <- utils.Result{
			Data: data.ID,
		}
	}()

	return output
}

func (c commandMongodbRepository) UpdateWallet(ctx context.Context, data walletModels.Wallet) <-chan utils.Result {
	output := make(chan utils.Result)

	go func() {
		defer close(output)
		objId, _ := primitive.ObjectIDFromHex(data.ID.String())
		err := c.mongoDb.UpdateOne(mongodb.UpdateOne{
			CollectionName: "wallet",
			Filter: bson.M{
				"userId": data.UserID,
				"_id":    objId,
			},
			Document: bson.M{
				"userId":         data.UserID,
				"balance":        data.Balance,
				"transactionLog": data.TransactionLog,
				"lastUpdated":    data.LastUpdated,
			},
		}, ctx)

		if err != nil {
			output <- utils.Result{
				Error: err,
			}
			return
		}

		output <- utils.Result{
			Data: data.ID,
		}
	}()

	return output
}
