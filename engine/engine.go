package engine

import (
	"fmt"
	"log"
	"time"
	"trading_bot/broker"
	"trading_bot/data"
	"trading_bot/models"
	"trading_bot/strategy"
)

const fixedQty = 50

type Engine struct {
	symbol      string
	timeframe   string
	strategy    strategy.Strategy
	dataSource  data.DataSource
	broker      broker.Broker
	position    *models.Position
	lastCandle  time.Time
	lastTradeAt time.Time
}

func New(symbol, timeframe string, str strategy.Strategy, ds data.DataSource, br broker.Broker) *Engine {
	return &Engine{
		symbol:     symbol,
		timeframe:  timeframe,
		strategy:   str,
		dataSource: ds,
		broker:     br,
	}
}

func (e *Engine) Run() error {
	candles, err := e.dataSource.FetchCandles(e.symbol, e.timeframe)
	if err != nil {
		return fmt.Errorf("fetch candles: %w", err)
	}
	if len(candles) == 0 {
		return fmt.Errorf("no candles returned")
	}

	latest := candles[len(candles)-1]
	if !e.lastCandle.IsZero() && !latest.Time.After(e.lastCandle) {
		log.Printf("no new candle available yet: %s", latest.Time.Format(time.RFC3339))
		return nil
	}
	e.lastCandle = latest.Time

	signal, atr := e.strategy.OnCandle(candles)
	log.Printf("candle=%s close=%.2f signal=%s atr=%.4f", latest.Time.Format(time.RFC3339), latest.Close, signal.String(), atr)

	if e.position != nil {
		if exited, err := e.maybeExit(latest, signal); err != nil {
			return err
		} else if exited {
			return nil
		}
		return nil
	}

	if signal == models.NoSignal || atr <= 0 {
		return nil
	}

	if !e.lastTradeAt.IsZero() && latest.Time.Equal(e.lastTradeAt) {
		log.Printf("trade skipped to avoid duplicate at candle=%s", latest.Time.Format(time.RFC3339))
		return nil
	}

	return e.enterPosition(latest, signal, atr)
}

func (e *Engine) maybeExit(c models.Candle, signal models.Signal) (bool, error) {
	pos := e.position
	if pos == nil {
		return false, nil
	}

	switch pos.Side {
	case models.SideLong:
		if c.High >= pos.Target {
			return true, e.closePosition(c.Time, pos.Target, "target_hit_long")
		}
		if c.Low <= pos.StopLoss {
			return true, e.closePosition(c.Time, pos.StopLoss, "stop_loss_long")
		}
		if signal == models.SellSignal {
			return true, e.closePosition(c.Time, c.Close, "reverse_signal")
		}
	case models.SideShort:
		if c.Low <= pos.Target {
			return true, e.closePosition(c.Time, pos.Target, "target_hit_short")
		}
		if c.High >= pos.StopLoss {
			return true, e.closePosition(c.Time, pos.StopLoss, "stop_loss_short")
		}
		if signal == models.BuySignal {
			return true, e.closePosition(c.Time, c.Close, "reverse_signal")
		}
	}

	return false, nil
}

func (e *Engine) enterPosition(c models.Candle, signal models.Signal, atr float64) error {
	order := models.Order{Symbol: e.symbol, Side: signal, Qty: fixedQty, Price: c.Close, OrderTag: "entry"}
	if err := e.broker.PlaceOrder(order); err != nil {
		return fmt.Errorf("place entry order: %w", err)
	}

	pos := &models.Position{
		Symbol:     e.symbol,
		Qty:        fixedQty,
		EntryPrice: c.Close,
		EntryTime:  c.Time,
	}
	if signal == models.BuySignal {
		pos.Side = models.SideLong
		pos.StopLoss = c.Close - (1.5 * atr)
		pos.Target = c.Close + (1.5 * atr)
	} else {
		pos.Side = models.SideShort
		pos.StopLoss = c.Close + (1.5 * atr)
		pos.Target = c.Close - (1.5 * atr)
	}

	e.position = pos
	e.lastTradeAt = c.Time
	log.Printf("entered %s position entry=%.2f stop_loss=%.2f target=%.2f atr=%.4f", pos.Side.String(), pos.EntryPrice, pos.StopLoss, pos.Target, atr)
	return nil
}

func (e *Engine) closePosition(ts time.Time, price float64, reason string) error {
	if e.position == nil {
		return nil
	}

	side := models.SellSignal
	if e.position.Side == models.SideShort {
		side = models.BuySignal
	}
	order := models.Order{Symbol: e.symbol, Side: side, Qty: e.position.Qty, Price: price, OrderTag: "exit_" + reason}
	if err := e.broker.PlaceOrder(order); err != nil {
		return fmt.Errorf("place exit order: %w", err)
	}

	pnl := 0.0
	if e.position.Side == models.SideLong {
		pnl = (price - e.position.EntryPrice) * float64(e.position.Qty)
	} else {
		pnl = (e.position.EntryPrice - price) * float64(e.position.Qty)
	}

	log.Printf("exited %s reason=%s exit_price=%.2f pnl=%.2f", e.position.Side.String(), reason, price, pnl)
	e.position = nil
	e.lastTradeAt = ts
	return nil
}
