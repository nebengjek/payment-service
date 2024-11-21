package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripOrderCompleted struct {
	OrderID        string  `json:"orderId" bson:"orderId"`
	RealDistance   float64 `json:"realDistance" bson:"realDistance"`
	FarePercentage float64 `json:"farePercentage" bson:"farePercentage"`
}

type User struct {
	Id           string `json:"_id" bson:"_id"`
	FullName     string `json:"fullName" bson:"fullName" validate:"required,min=3,max=100"`
	MobileNumber string `json:"mobileNumber" bson:"mobileNumber" validate:"required"`
}

type BillingResponse struct {
	OrderID     string  `json:"orderId"`
	PassengerID string  `json:"passengerId"`
	Trip        Trip    `json:"trip"`
	Driver      User    `json:"driver"`
	TotalFare   float64 `json:"totalFare"`
	Status      string  `json:"status"`
}

type Trip struct {
	Origin      Location `json:"origin"`
	Destination Location `json:"destination"`
	DistanceKm  float64  `json:"distanceKm"`
}

type TripOrder struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	OrderID       string             `json:"orderId" bson:"orderId"`
	PassengerID   string             `json:"passengerId" bson:"passengerId"`
	DriverID      string             `json:"driverId,omitempty" bson:"driverId,omitempty"`
	Origin        Location           `json:"origin" bson:"origin"`
	Destination   Location           `json:"destination" bson:"destination"`
	Status        string             `json:"status" bson:"status"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	EstimatedFare float64            `json:"estimatedFare" bson:"estimatedFare"`
	DistanceKm    float64            `json:"distanceKm" bson:"distanceKm"`
	RealDistance  float64            `json:"realDistance,omitempty" bson:"realDistance,omitempty"`
}

type Location struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	Address   string  `json:"address" bson:"address"`
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
