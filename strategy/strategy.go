package strategy

import "trading_bot/models"

type Strategy interface {
	Name() string
	OnCandle(candles []models.Candle) (models.Signal, float64)
}
