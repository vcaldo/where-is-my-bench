package storage

import (
	"context"

	"github.com/vcaldo/where-is-my-bench/telegram-bot/pkg/bench"
)

type BenchStorage interface {
	StoreBenches(ctx context.Context, benches []bench.Bench) error
	FindNearby(ctx context.Context, lat, lon float64, radiusMeters float64) ([]bench.Bench, error)
}
