package wallet

import (
	"context"

	"payment-service/bin/modules/wallet/models"
	"payment-service/bin/pkg/utils"
)

type UsecaseQuery interface {
	// idiomatic go, ctx first before payload. See https://pkg.go.dev/context#pkg-overview
}

type UsecaseCommand interface {
	// idiomatic go, ctx first before payload. See https://pkg.go.dev/context#pkg-overview
	TopUpWallet(ctx context.Context, payload models.TopUpRequest) utils.Result
}

type MongodbRepositoryQuery interface {
	FindUser(ctx context.Context, userId string) <-chan utils.Result
	Findwallet(ctx context.Context, userId string) <-chan utils.Result
}

type MongodbRepositoryCommand interface {
	NewObjectID(ctx context.Context) string
	Insertwallet(ctx context.Context, data models.Wallet) <-chan utils.Result
	UpdateWallet(ctx context.Context, data models.Wallet) <-chan utils.Result
}
