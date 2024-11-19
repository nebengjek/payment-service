package queries

import (
	"context"
	driver "payment-service/bin/modules/billing"
	"payment-service/bin/modules/billing/models"
	"payment-service/bin/pkg/databases/mongodb"
	"payment-service/bin/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type queryMongodbRepository struct {
	mongoDb mongodb.MongoDBLogger
}

func NewQueryMongodbRepository(mongodb mongodb.MongoDBLogger) driver.MongodbRepositoryQuery {
	return &queryMongodbRepository{
		mongoDb: mongodb,
	}
}

func (q queryMongodbRepository) FindDriver(ctx context.Context, userId string) <-chan utils.Result {
	output := make(chan utils.Result)

	go func() {
		defer close(output)
		objId, _ := primitive.ObjectIDFromHex(userId)

		var driver models.User
		err := q.mongoDb.FindOne(mongodb.FindOne{
			Result:         &driver,
			CollectionName: "user",
			Filter: bson.M{
				"_id": objId,
			},
		}, ctx)
		if err != nil {
			output <- utils.Result{
				Error: err,
			}
		}

		output <- utils.Result{
			Data: driver,
		}

	}()

	return output
}

func (q queryMongodbRepository) FindActiveOrderPassanger(ctx context.Context, orderId string) <-chan utils.Result {
	output := make(chan utils.Result)
	go func() {
		defer close(output)
		var trip models.TripOrder
		err := q.mongoDb.FindOne(mongodb.FindOne{
			Result:         &trip,
			CollectionName: "trip-orders",
			Filter: bson.M{
				"orderId": orderId,
			},
		}, ctx)
		if err != nil {
			output <- utils.Result{
				Error: err,
			}
		}
		output <- utils.Result{
			Data: trip,
		}

	}()

	return output
}
