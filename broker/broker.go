package broker

import "trading_bot/models"

type Broker interface {
	PlaceOrder(order models.Order) error
}
