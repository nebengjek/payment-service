package handlers

import (
	"payment-service/bin/middlewares"
	billing "payment-service/bin/modules/billing"

	"payment-service/bin/pkg/utils"

	"github.com/labstack/echo/v4"
)

type billingHttpHandler struct {
	billingUsecaseQuery   billing.UsecaseQuery
	billingUseCaseCommand billing.UsecaseCommand
}

func InitbillingHttpHandler(e *echo.Echo, uq billing.UsecaseQuery, uc billing.UsecaseCommand) {

	handler := &billingHttpHandler{
		billingUsecaseQuery:   uq,
		billingUseCaseCommand: uc,
	}
	route := e.Group("/billing")
	route.GET("/v1/trip-bill/:orderId", handler.BillingTrip, middlewares.VerifyBearer)
}

func (u billingHttpHandler) BillingTrip(c echo.Context) error {
	orderId := c.Param("orderId")
	userId := utils.ConvertString(c.Get("userId"))
	result := u.billingUsecaseQuery.BillingTrip(c.Request().Context(), userId, orderId)

	if result.Error != nil {
		return utils.ResponseError(result.Error, c)
	}

	return utils.Response(result.Data, "Get Billing Trip", 200, c)
}
