package broker

import (
	"log"
	"trading_bot/models"
)

type MockBroker struct{}

func NewMockBroker() *MockBroker {
	return &MockBroker{}
}

func (m *MockBroker) PlaceOrder(order models.Order) error {
	log.Printf("[MOCK BROKER] placing order side=%s symbol=%s qty=%d price=%.2f tag=%s", order.Side.String(), order.Symbol, order.Qty, order.Price, order.OrderTag)
	return nil
}
