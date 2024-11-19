package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	billing "payment-service/bin/modules/billing"
	"payment-service/bin/modules/billing/models"
	kafkaPkgConfluent "payment-service/bin/pkg/kafka/confluent"
	"payment-service/bin/pkg/log"

	k "gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type billingHandler struct {
	billingUsecaseCommand billing.UsecaseCommand
}

func NewBillingConsumer(dc billing.UsecaseCommand) kafkaPkgConfluent.ConsumerHandler {
	return &billingHandler{
		billingUsecaseCommand: dc,
	}
}

func (i billingHandler) HandleMessage(message *k.Message) {
	log.GetLogger().Info("consumer", fmt.Sprintf("Partition: %v - Offset: %v", message.TopicPartition.Partition, message.TopicPartition.Offset.String()), *message.TopicPartition.Topic, string(message.Value))

	var msg models.TripOrderCompleted
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		log.GetLogger().Error("consumer", "unmarshal-data", err.Error(), string(message.Value))
		return
	}

	if err := i.billingUsecaseCommand.CreateBilling(context.Background(), msg); err != nil {
		log.GetLogger().Error("consumer", "CreateBilling", err.Error(), string(message.Value))
		return
	}

	return
}
