package commands

import (
	"context"

	user "payment-service/bin/modules/wallet"
	"payment-service/bin/modules/wallet/models"
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

func (c commandMongodbRepository) Insertwallet(ctx context.Context, data models.Wallet) <-chan utils.Result {
	output := make(chan utils.Result)

	go func() {
		defer close(output)

		err := c.mongoDb.InsertOne(mongodb.InsertOne{
			CollectionName: "wallet",
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

func (c commandMongodbRepository) UpdateWallet(ctx context.Context, data models.Wallet) <-chan utils.Result {
	output := make(chan utils.Result)

	go func() {
		defer close(output)
		objId, _ := primitive.ObjectIDFromHex(data.ID.String())
		err := c.mongoDb.UpsertOne(mongodb.UpsertOne{
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
