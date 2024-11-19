package commands

import (
	"context"

	user "payment-service/bin/modules/billing"
	"payment-service/bin/modules/billing/models"
	"payment-service/bin/pkg/databases/mongodb"
	"payment-service/bin/pkg/utils"

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

		err := c.mongoDb.InsertOne(mongodb.InsertOne{
			CollectionName: "billing",
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
