package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"trading_bot/broker"
	"trading_bot/config"
	"trading_bot/data"
	"trading_bot/engine"
	"trading_bot/strategy"
)

func main() {
	cfgPath := flag.String("config", "config.json", "path to JSON config")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	str, err := strategyByName(cfg.Strategy)
	if err != nil {
		log.Fatalf("strategy setup failed: %v", err)
	}

	ds, err := dataSourceByName(cfg.DataSource)
	if err != nil {
		log.Fatalf("data source setup failed: %v", err)
	}

	br, err := brokerByName(cfg)
	if err != nil {
		log.Fatalf("broker setup failed: %v", err)
	}

	e := engine.New(cfg.Symbol, cfg.Timeframe, str, ds, br)
	interval, err := time.ParseDuration(cfg.Timeframe)
	if err != nil {
		log.Fatalf("invalid timeframe: %v", err)
	}

	log.Printf("starting trading bot strategy=%s symbol=%s timeframe=%s datasource=%s broker=%s", cfg.Strategy, cfg.Symbol, cfg.Timeframe, cfg.DataSource, cfg.Broker)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := e.Run(); err != nil {
		log.Printf("engine run error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := e.Run(); err != nil {
				log.Printf("engine run error: %v", err)
			}
		case sig := <-sigCh:
			log.Printf("shutting down trading bot on signal: %s", sig.String())
			return
		}
	}
}

func strategyByName(name string) (strategy.Strategy, error) {
	switch name {
	case "ema_rsi_atr":
		return strategy.NewEMARSIATRStrategy(), nil
	default:
		return nil, errUnsupported("strategy", name)
	}
}

func dataSourceByName(name string) (data.DataSource, error) {
	switch name {
	case "mock":
		return data.NewMockDataSource(time.Now().UnixNano()), nil
	case "websocket":
		// TODO: Wire live websocket implementation once provider integration is completed.
		return data.NewWebSocketDataSource(), nil
	default:
		return nil, errUnsupported("data source", name)
	}
}

func brokerByName(cfg config.Config) (broker.Broker, error) {
	switch cfg.Broker {
	case "mock":
		return broker.NewMockBroker(), nil
	case "zerodha":
		zAuth := broker.ZerodhaAuth{APIKey: cfg.Zerodha.APIKey, AccessToken: cfg.Zerodha.AccessToken}
		return broker.NewZerodhaBroker(&http.Client{Timeout: 10 * time.Second}, zAuth), nil
	default:
		return nil, errUnsupported("broker", cfg.Broker)
	}
}

func errUnsupported(component, name string) error {
	return &unsupportedError{component: component, name: name}
}

type unsupportedError struct {
	component string
	name      string
}

func (e *unsupportedError) Error() string {
	return "unsupported " + e.component + ": " + e.name
}
