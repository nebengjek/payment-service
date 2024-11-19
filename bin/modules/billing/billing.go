package billing

import (
	"context"

	"payment-service/bin/modules/billing/models"
	"payment-service/bin/pkg/utils"
)

type UsecaseQuery interface {
}

type UsecaseCommand interface {
	CreateBilling(ctx context.Context, payload models.TripOrderCompleted) error
}

type MongodbRepositoryQuery interface {
	FindDriver(ctx context.Context, userId string) <-chan utils.Result
	FindActiveOrderPassanger(ctx context.Context, orderId string) <-chan utils.Result
}

type MongodbRepositoryCommand interface {
	NewObjectID(ctx context.Context) string
	InsertBilling(ctx context.Context, data models.Transaction) <-chan utils.Result
}
