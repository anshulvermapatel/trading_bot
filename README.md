# Trading Bot (Go 1.22+)

Production-ready algorithmic trading foundation with modular architecture:

- `strategy/` strategy implementations
- `engine/` execution and risk management
- `broker/` broker adapters (mock + Zerodha scaffold)
- `data/` market data providers (mock + websocket placeholder)
- `models/` shared domain models
- `config/` config loader
- `main.go` entrypoint and dependency wiring

## Strategy implemented

`ema_rsi_atr`:

- EMA 21 + EMA 50 trend filter
- RSI (14) momentum filter
- ATR (14) risk metric
- Fresh signal logic (true now, false previous candle)
- Time windows (IST): 09:30–11:30 and 13:00–15:15
- One position at a time
- Executes on candle close

## Risk Management

Fixed ATR stop loss set **at entry** (not trailing):

- Long SL = entry - (1.5 × ATR)
- Short SL = entry + (1.5 × ATR)
- Long target = entry + (1.5 × ATR)
- Short target = entry - (1.5 × ATR)

Engine exits a position when either fixed stop loss or fixed target is hit.

## Run

```bash
go run . -config config.json
```

## Notes

- `mock` data source continuously synthesizes candles for local testing.
- `websocket` data source is intentionally a TODO placeholder.
- `zerodha` broker has REST order placement scaffold for Kite Connect.

## TODOs

1. Implement full Zerodha session/auth lifecycle (request token exchange, refresh flow, secure secret storage).
2. Parse and persist Kite order responses and errors for auditing.
3. Implement live websocket market data ingestion and candle builder.
4. Add persistent state/recovery for open positions across restarts.
