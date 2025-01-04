package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/pkg/bench"
)

const benchesKey = "benches"

type BenchStore struct {
	rdb *redis.Client
}

func NewBenchStore(addr string, password string, db int) *BenchStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &BenchStore{rdb: rdb}
}

func (s *BenchStore) StoreBenches(ctx context.Context, benches []bench.Bench) error {
	pipe := s.rdb.Pipeline()
	for _, b := range benches {
		pipe.GeoAdd(ctx, benchesKey, &redis.GeoLocation{
			Name:      b.GisID,
			Longitude: b.Longitude,
			Latitude:  b.Latitude,
		})
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *BenchStore) FindNearby(ctx context.Context, lat, lon float64, radiusMeters float64) ([]bench.Bench, error) {
	locs, err := s.rdb.GeoRadius(ctx, benchesKey, lon, lat, &redis.GeoRadiusQuery{
		Radius: radiusMeters,
		Unit:   "m",
		Sort:   "ASC",
	}).Result()
	if err != nil {
		return nil, err
	}
	fmt.Printf("found %d benches\n", len(locs))

	benches := make([]bench.Bench, len(locs))
	for i, loc := range locs {
		benches[i] = bench.Bench{
			GisID:     loc.Name,
			Longitude: loc.Longitude,
			Latitude:  loc.Latitude,
		}
	}
	return benches, nil
}
