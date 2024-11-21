package usecases

import (
	"context"
	"fmt"
	billing "payment-service/bin/modules/billing"
	"payment-service/bin/modules/billing/models"
	httpError "payment-service/bin/pkg/http-error"
	"payment-service/bin/pkg/log"
	"payment-service/bin/pkg/utils"

	"github.com/redis/go-redis/v9"
)

type queryUsecase struct {
	billingRepositoryQuery billing.MongodbRepositoryQuery
	redisClient            redis.UniversalClient
}

func NewQueryUsecase(mq billing.MongodbRepositoryQuery, rh redis.UniversalClient) billing.UsecaseQuery {
	return &queryUsecase{
		billingRepositoryQuery: mq,
		redisClient:            rh,
	}
}

func (q *queryUsecase) BillingTrip(ctx context.Context, userId string, orderId string) utils.Result {
	var result utils.Result
	trip := <-q.billingRepositoryQuery.FindActiveOrderPassanger(ctx, orderId)
	if trip.Error != nil {
		errObj := httpError.NewNotFound()
		errObj.Message = fmt.Sprintf("The order has Notfound: %v", trip.Error)
		log.GetLogger().Error("command_usecase", "The order has Notfound", "CompletedTrip", utils.ConvertString(trip.Error))
		result.Error = errObj
		return result
	}
	billing := <-q.billingRepositoryQuery.FindBillingPassanger(ctx, userId, orderId)
	if billing.Error != nil {
		errObj := httpError.NewNotFound()
		errObj.Message = fmt.Sprintf("The billing has Notfound: %v", billing.Error)
		log.GetLogger().Error("command_usecase", "The order has Notfound", "CompletedTrip", utils.ConvertString(billing.Error))
		result.Error = errObj
		return result
	}
	tripOrder := trip.Data.(models.TripOrder)
	billingTrip := billing.Data.(models.Transaction)
	driver := <-q.billingRepositoryQuery.FindDriver(ctx, tripOrder.DriverID)
	if driver.Error != nil {
		errObj := httpError.NewNotFound()
		errObj.Message = fmt.Sprintf("profile driver has Notfound: %v", driver.Error)
		log.GetLogger().Error("command_usecase", "profile driver has Notfound", "CompletedTrip", utils.ConvertString(driver.Error))
		result.Error = errObj
		return result
	}
	infoDriver := driver.Data.(models.User)
	result.Data = models.BillingResponse{
		OrderID:     orderId,
		PassengerID: userId,
		Trip: models.Trip{
			Origin:      tripOrder.Origin,
			Destination: tripOrder.Destination,
			DistanceKm:  tripOrder.RealDistance,
		},
		Driver:    infoDriver,
		TotalFare: billingTrip.TotalFare,
		Status:    billingTrip.Status,
	}
	return result
}
