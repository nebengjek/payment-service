package handlers

import (
	billing "payment-service/bin/modules/billing"
	kafkaPkgConfluent "payment-service/bin/pkg/kafka/confluent"
)

func InitPaymentEventHandler(billing billing.UsecaseCommand, kc kafkaPkgConfluent.Consumer) {

	kc.SetHandler(NewBillingConsumer(billing))
	kc.Subscribe("create-billing")

}
