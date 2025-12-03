//go:build js && wasm

package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/gaarutyunov/guix/pkg/runtime/chart"
)

// MarkovChainGenerator generates realistic OHLCV data using Markov chains
type MarkovChainGenerator struct {
	currentPrice  float64
	currentVolume float64
	volatility    float64
	trendBias     float64
	rand          *rand.Rand
}

// NewMarkovChainGenerator creates a new Markov chain-based data generator
func NewMarkovChainGenerator(startPrice, volatility float64) *MarkovChainGenerator {
	seed := time.Now().UnixNano()
	return &MarkovChainGenerator{
		currentPrice:  startPrice,
		currentVolume: 1000000000, // Start with 1B volume
		volatility:    volatility,
		trendBias:     0.0,
		rand:          rand.New(rand.NewSource(seed)),
	}
}

// GenerateCandles generates n candles using Markov chain transitions
func (mcg *MarkovChainGenerator) GenerateCandles(count int, interval string) []chart.OHLCV {
	candles := make([]chart.OHLCV, count)
	intervalMs := getIntervalMilliseconds(interval)
	startTime := time.Now().Add(-time.Duration(count*int(intervalMs)) * time.Millisecond).UnixMilli()

	for i := 0; i < count; i++ {
		timestamp := startTime + int64(i)*intervalMs
		candle := mcg.generateNextCandle(timestamp)
		candles[i] = candle

		// Occasionally change trend bias (simulate market regime changes)
		if mcg.rand.Float64() < 0.05 { // 5% chance per candle
			mcg.trendBias = (mcg.rand.Float64() - 0.5) * 0.002 // -0.1% to +0.1% bias
		}
	}

	return candles
}

// generateNextCandle generates a single candle using Markov state transitions
func (mcg *MarkovChainGenerator) generateNextCandle(timestamp int64) chart.OHLCV {
	// Determine candle direction based on Markov state
	// States: uptrend (bullish), downtrend (bearish), ranging
	state := mcg.getMarkovState()

	// Generate price movement based on state
	var priceChange float64
	switch state {
	case "uptrend":
		// Bullish candle with higher probability
		priceChange = (mcg.rand.Float64()*0.02 + mcg.trendBias) * mcg.currentPrice
		if mcg.rand.Float64() < 0.7 { // 70% chance of up candle
			priceChange = math.Abs(priceChange)
		} else {
			priceChange = -math.Abs(priceChange) * 0.5 // Smaller down moves
		}
	case "downtrend":
		// Bearish candle with higher probability
		priceChange = (mcg.rand.Float64()*0.02 - mcg.trendBias) * mcg.currentPrice
		if mcg.rand.Float64() < 0.7 { // 70% chance of down candle
			priceChange = -math.Abs(priceChange)
		} else {
			priceChange = math.Abs(priceChange) * 0.5 // Smaller up moves
		}
	default: // ranging
		// Random walk with smaller moves
		priceChange = (mcg.rand.Float64() - 0.5) * 0.01 * mcg.currentPrice
	}

	// Apply volatility
	priceChange *= mcg.volatility

	// Generate OHLC based on price change
	open := mcg.currentPrice
	close := open + priceChange

	// Generate realistic high/low wicks
	wickSize := math.Abs(priceChange) * (0.5 + mcg.rand.Float64()*1.5)
	high := math.Max(open, close) + wickSize*mcg.rand.Float64()
	low := math.Min(open, close) - wickSize*mcg.rand.Float64()

	// Ensure OHLC relationships are valid
	if high < math.Max(open, close) {
		high = math.Max(open, close) * 1.001
	}
	if low > math.Min(open, close) {
		low = math.Min(open, close) * 0.999
	}

	// Generate volume with some correlation to price movement
	volumeChange := (mcg.rand.Float64() - 0.5) * 0.3 // ±30% volume change
	volumeChange += math.Abs(priceChange/open) * 2   // Higher volume on bigger moves
	mcg.currentVolume *= (1 + volumeChange)

	// Clamp volume to reasonable range
	if mcg.currentVolume < 100000000 { // Min 100M
		mcg.currentVolume = 100000000
	}
	if mcg.currentVolume > 10000000000 { // Max 10B
		mcg.currentVolume = 10000000000
	}

	// Update current price for next iteration
	mcg.currentPrice = close

	return chart.OHLCV{
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    mcg.currentVolume,
	}
}

// getMarkovState determines the current market state using Markov chain logic
func (mcg *MarkovChainGenerator) getMarkovState() string {
	// Transition probabilities based on trend bias
	rand := mcg.rand.Float64()

	if mcg.trendBias > 0.0005 {
		// In uptrend
		if rand < 0.6 { // 60% stay in uptrend
			return "uptrend"
		} else if rand < 0.85 { // 25% transition to ranging
			mcg.trendBias *= 0.5
			return "ranging"
		} else { // 15% transition to downtrend
			mcg.trendBias = -math.Abs(mcg.trendBias)
			return "downtrend"
		}
	} else if mcg.trendBias < -0.0005 {
		// In downtrend
		if rand < 0.6 { // 60% stay in downtrend
			return "downtrend"
		} else if rand < 0.85 { // 25% transition to ranging
			mcg.trendBias *= 0.5
			return "ranging"
		} else { // 15% transition to uptrend
			mcg.trendBias = math.Abs(mcg.trendBias)
			return "uptrend"
		}
	} else {
		// In ranging state
		if rand < 0.5 { // 50% stay in ranging
			return "ranging"
		} else if rand < 0.75 { // 25% transition to uptrend
			mcg.trendBias = 0.001
			return "uptrend"
		} else { // 25% transition to downtrend
			mcg.trendBias = -0.001
			return "downtrend"
		}
	}
}

// getIntervalMilliseconds converts interval string to milliseconds
func getIntervalMilliseconds(interval string) int64 {
	switch interval {
	case "1m":
		return 60 * 1000
	case "5m":
		return 5 * 60 * 1000
	case "15m":
		return 15 * 60 * 1000
	case "30m":
		return 30 * 60 * 1000
	case "1h":
		return 60 * 60 * 1000
	case "4h":
		return 4 * 60 * 60 * 1000
	case "1d":
		return 24 * 60 * 60 * 1000
	default:
		return 60 * 60 * 1000 // Default to 1h
	}
}

// GenerateFallbackData generates realistic fallback data using Markov chains
func GenerateFallbackData(count int) []chart.OHLCV {
	// Start with Bitcoin's approximate current price
	generator := NewMarkovChainGenerator(45000.0, 1.0)
	return generator.GenerateCandles(count, "1h")
}
