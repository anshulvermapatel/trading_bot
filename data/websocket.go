package data

import (
	"errors"
	"trading_bot/models"
)

type WebSocketDataSource struct {
	// TODO: add websocket client + subscription handling for live candle aggregation.
}

func NewWebSocketDataSource() *WebSocketDataSource {
	return &WebSocketDataSource{}
}

func (w *WebSocketDataSource) FetchCandles(symbol string, timeframe string) ([]models.Candle, error) {
	_ = symbol
	_ = timeframe
	return nil, errors.New("websocket datasource not implemented yet")
}
