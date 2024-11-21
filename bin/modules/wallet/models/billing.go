package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id           string `json:"_id" bson:"_id"`
	FullName     string `json:"fullName" bson:"fullName" validate:"required,min=3,max=100"`
	MobileNumber string `json:"mobileNumber" bson:"mobileNumber" validate:"required"`
}

type TopUpRequest struct {
	UserID string  `json:"userId"`
	Amount float64 `json:"amount"`
}

type TopUpResponse struct {
	Message string  `json:"message"`
	Balance float64 `json:"balance"`
}

type Wallet struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         string             `bson:"userId" json:"userId"`
	Balance        float64            `bson:"balance" json:"balance"`
	TransactionLog []TransactionLog   `bson:"transactionLog" json:"transactionLog"`
	LastUpdated    time.Time          `bson:"lastUpdated" json:"lastUpdated"`
}

type TransactionLog struct {
	TransactionID string    `bson:"transactionId" json:"transactionId"`
	Amount        float64   `bson:"amount" json:"amount"`
	Type          string    `bson:"type" json:"type"`
	Description   string    `bson:"description" json:"description"`
	Timestamp     time.Time `bson:"timestamp" json:"timestamp"`
}

type Transaction struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	OrderID        string             `json:"orderId" bson:"orderId"`
	PassengerID    string             `json:"passengerId" bson:"passengerId"`
	DriverID       string             `json:"driverId" bson:"driverId"`
	TotalFare      float64            `json:"totalFare" bson:"totalFare"`
	AdminFee       float64            `json:"adminFee" bson:"adminFee"`
	DriverEarnings float64            `json:"driverEarnings" bson:"driverEarnings"`
	PaymentMethod  string             `json:"paymentMethod" bson:"paymentMethod"`
	Status         string             `json:"status" bson:"status"`
	Timestamp      time.Time          `json:"timestamp" bson:"timestamp"`
}
