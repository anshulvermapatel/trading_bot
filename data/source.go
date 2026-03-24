package data

import "trading_bot/models"

type DataSource interface {
	FetchCandles(symbol string, timeframe string) ([]models.Candle, error)
}
