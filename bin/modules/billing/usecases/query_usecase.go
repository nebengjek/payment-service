package usecases

import (
	driver "payment-service/bin/modules/billing"

	"github.com/redis/go-redis/v9"
)

type queryUsecase struct {
	driverRepositoryQuery driver.MongodbRepositoryQuery
	redisClient           redis.UniversalClient
}

func NewQueryUsecase(mq driver.MongodbRepositoryQuery, rh redis.UniversalClient) driver.UsecaseQuery {
	return &queryUsecase{
		driverRepositoryQuery: mq,
		redisClient:           rh,
	}
}
