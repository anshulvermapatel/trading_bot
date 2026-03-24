package strategy

import (
	"log"
	"math"
	"time"
	"trading_bot/models"
)

const (
	emaFastPeriod = 21
	emaSlowPeriod = 50
	rsiPeriod     = 14
	atrPeriod     = 14
)

type EMARSIATRStrategy struct {
	loc *time.Location
}

func NewEMARSIATRStrategy() *EMARSIATRStrategy {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Printf("failed to load Asia/Kolkata timezone, using UTC: %v", err)
		loc = time.UTC
	}
	return &EMARSIATRStrategy{loc: loc}
}

func (s *EMARSIATRStrategy) Name() string {
	return "ema_rsi_atr"
}

func (s *EMARSIATRStrategy) OnCandle(candles []models.Candle) (models.Signal, float64) {
	minCandles := emaSlowPeriod + rsiPeriod + 2
	if len(candles) < minCandles {
		return models.NoSignal, 0
	}

	atr := calcATR(candles, atrPeriod)
	if atr <= 0 {
		return models.NoSignal, 0
	}

	latest := candles[len(candles)-1]
	if !s.inTradeWindow(latest.Time) {
		return models.NoSignal, atr
	}

	closesNow := closes(candles)
	closesPrev := closes(candles[:len(candles)-1])

	ema21Now := calcEMA(closesNow, emaFastPeriod)
	ema50Now := calcEMA(closesNow, emaSlowPeriod)
	rsiNow := calcRSI(closesNow, rsiPeriod)

	ema21Prev := calcEMA(closesPrev, emaFastPeriod)
	ema50Prev := calcEMA(closesPrev, emaSlowPeriod)
	rsiPrev := calcRSI(closesPrev, rsiPeriod)

	buyNow := ema21Now > ema50Now && rsiNow > 55
	buyPrev := ema21Prev > ema50Prev && rsiPrev > 55
	sellNow := ema21Now < ema50Now && rsiNow < 45
	sellPrev := ema21Prev < ema50Prev && rsiPrev < 45

	switch {
	case buyNow && !buyPrev:
		return models.BuySignal, atr
	case sellNow && !sellPrev:
		return models.SellSignal, atr
	default:
		return models.NoSignal, atr
	}
}

func (s *EMARSIATRStrategy) inTradeWindow(t time.Time) bool {
	local := t.In(s.loc)
	minutes := local.Hour()*60 + local.Minute()

	mornStart := 9*60 + 30
	mornEnd := 11*60 + 30
	aftStart := 13 * 60
	aftEnd := 15*60 + 15

	return (minutes >= mornStart && minutes <= mornEnd) ||
		(minutes >= aftStart && minutes <= aftEnd)
}

func closes(candles []models.Candle) []float64 {
	out := make([]float64, len(candles))
	for i := range candles {
		out[i] = candles[i].Close
	}
	return out
}

func calcEMA(values []float64, period int) float64 {
	if len(values) < period || period <= 0 {
		return 0
	}

	k := 2.0 / float64(period+1)
	ema := avg(values[:period])
	for i := period; i < len(values); i++ {
		ema = values[i]*k + ema*(1-k)
	}
	return ema
}

func calcRSI(closes []float64, period int) float64 {
	if len(closes) <= period {
		return 50
	}

	gain := 0.0
	loss := 0.0
	for i := 1; i <= period; i++ {
		delta := closes[i] - closes[i-1]
		if delta > 0 {
			gain += delta
		} else {
			loss -= delta
		}
	}
	avgGain := gain / float64(period)
	avgLoss := loss / float64(period)

	for i := period + 1; i < len(closes); i++ {
		delta := closes[i] - closes[i-1]
		currentGain := math.Max(delta, 0)
		currentLoss := math.Max(-delta, 0)
		avgGain = ((avgGain * float64(period-1)) + currentGain) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + currentLoss) / float64(period)
	}

	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

func calcATR(candles []models.Candle, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	trs := make([]float64, 0, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close
		tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		trs = append(trs, tr)
	}

	atr := avg(trs[:period])
	for i := period; i < len(trs); i++ {
		atr = ((atr * float64(period-1)) + trs[i]) / float64(period)
	}
	return atr
}

func avg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
