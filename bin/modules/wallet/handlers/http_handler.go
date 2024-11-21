package handlers

import (
	"payment-service/bin/middlewares"
	wallet "payment-service/bin/modules/wallet"
	"payment-service/bin/modules/wallet/models"

	"payment-service/bin/pkg/utils"

	"github.com/labstack/echo/v4"
)

type walletHttpHandler struct {
	walletUsecaseQuery   wallet.UsecaseQuery
	walletUseCaseCommand wallet.UsecaseCommand
}

func InitwalletHttpHandler(e *echo.Echo, uq wallet.UsecaseQuery, uc wallet.UsecaseCommand) {

	handler := &walletHttpHandler{
		walletUsecaseQuery:   uq,
		walletUseCaseCommand: uc,
	}
	route := e.Group("/wallet")
	route.POST("/v1/topup", handler.TopUpWallet, middlewares.VerifyBearer)
	route.POST("/v1/approve", handler.TopUpWallet, middlewares.VerifyBearer)
}

func (u walletHttpHandler) TopUpWallet(c echo.Context) error {
	req := new(models.TopUpRequest)

	if err := c.Bind(&req); err != nil {
		return utils.ResponseError(err, c)
	}

	if err := c.Validate(req); err != nil {
		return utils.ResponseError(err, c)
	}

	userId := utils.ConvertString(c.Get("userId"))
	result := u.walletUseCaseCommand.TopUpWallet(c.Request().Context(), userId, *req)

	if result.Error != nil {
		return utils.ResponseError(result.Error, c)
	}

	return utils.Response(result.Data, "Top up Wallet", 200, c)
}
