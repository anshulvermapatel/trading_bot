package data

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
	"trading_bot/models"
)

type MockDataSource struct {
	mu      sync.Mutex
	history map[string][]models.Candle
	rand    *rand.Rand
}

func NewMockDataSource(seed int64) *MockDataSource {
	return &MockDataSource{
		history: make(map[string][]models.Candle),
		rand:    rand.New(rand.NewSource(seed)),
	}
}

func (m *MockDataSource) FetchCandles(symbol string, timeframe string) ([]models.Candle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, err := parseTimeframe(timeframe)
	if err != nil {
		return nil, err
	}

	candles, ok := m.history[symbol]
	if !ok {
		candles = m.seedCandles(d)
	} else {
		candles = append(candles, m.nextCandle(candles[len(candles)-1], d))
		if len(candles) > 600 {
			candles = candles[len(candles)-600:]
		}
	}
	m.history[symbol] = candles
	copyOut := make([]models.Candle, len(candles))
	copy(copyOut, candles)
	return copyOut, nil
}

func (m *MockDataSource) seedCandles(step time.Duration) []models.Candle {
	now := time.Now().UTC().Truncate(step)
	candles := make([]models.Candle, 0, 260)
	price := 22000.0
	for i := 260; i > 0; i-- {
		t := now.Add(-time.Duration(i) * step)
		open := price
		drift := m.rand.NormFloat64() * 12
		close := open + drift
		high := max(open, close) + m.rand.Float64()*4
		low := min(open, close) - m.rand.Float64()*4
		candles = append(candles, models.Candle{Time: t, Open: open, High: high, Low: low, Close: close, Volume: 1000 + m.rand.Float64()*5000})
		price = close
	}
	return candles
}

func (m *MockDataSource) nextCandle(prev models.Candle, step time.Duration) models.Candle {
	t := prev.Time.Add(step)
	open := prev.Close
	drift := m.rand.NormFloat64() * 10
	close := open + drift
	high := max(open, close) + m.rand.Float64()*3
	low := min(open, close) - m.rand.Float64()*3
	return models.Candle{Time: t, Open: open, High: high, Low: low, Close: close, Volume: 1000 + m.rand.Float64()*5000}
}

func parseTimeframe(tf string) (time.Duration, error) {
	tf = strings.TrimSpace(strings.ToLower(tf))
	switch tf {
	case "1m", "3m", "5m", "15m", "30m", "1h", "4h":
		return time.ParseDuration(tf)
	default:
		return 0, fmt.Errorf("unsupported timeframe: %s", tf)
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
