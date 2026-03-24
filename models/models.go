package models

import "time"

type Signal int

const (
	NoSignal Signal = iota
	BuySignal
	SellSignal
)

func (s Signal) String() string {
	switch s {
	case BuySignal:
		return "BUY"
	case SellSignal:
		return "SELL"
	default:
		return "NO_SIGNAL"
	}
}

type Side int

const (
	SideFlat Side = iota
	SideLong
	SideShort
)

func (s Side) String() string {
	switch s {
	case SideLong:
		return "LONG"
	case SideShort:
		return "SHORT"
	default:
		return "FLAT"
	}
}

type Candle struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume"`
}

type Position struct {
	Symbol     string
	Side       Side
	Qty        int
	EntryPrice float64
	StopLoss   float64
	Target     float64
	EntryTime  time.Time
}

type Order struct {
	Symbol   string
	Side     Signal
	Qty      int
	Price    float64
	OrderTag string
}
