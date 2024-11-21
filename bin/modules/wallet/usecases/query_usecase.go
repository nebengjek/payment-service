package usecases

import (
	wallet "payment-service/bin/modules/wallet"

	"github.com/redis/go-redis/v9"
)

type queryUsecase struct {
	walletRepositoryQuery wallet.MongodbRepositoryQuery
	redisClient           redis.UniversalClient
}

func NewQueryUsecase(mq wallet.MongodbRepositoryQuery, rh redis.UniversalClient) wallet.UsecaseQuery {
	return &queryUsecase{
		walletRepositoryQuery: mq,
		redisClient:           rh,
	}
}
