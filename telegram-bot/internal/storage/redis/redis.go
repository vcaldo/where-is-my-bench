package redis

import (
	"context"
	"fmt"
	"strconv"

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
		// Store geospatial data
		pipe.GeoAdd(ctx, benchesKey, &redis.GeoLocation{
			Name:      b.GisID,
			Longitude: b.Longitude,
			Latitude:  b.Latitude,
		})

		// Store complete bench data in hash
		benchKey := fmt.Sprintf("bench:%s", b.GisID)
		pipe.HSet(ctx, benchKey, map[string]interface{}{
			"type":              b.Type,
			"code":              b.Code,
			"description":       b.Description,
			"manufacturer":      b.Manufacturer,
			"district_code":     b.DistrictCode,
			"district_name":     b.DistrictName,
			"neighborhood_code": b.NeighborhoodCode,
			"neighborhood_name": b.NeighborhoodName,
			"zone":              b.Zone,
			"street_name":       b.StreetName,
			"street_number":     b.StreetNumber,
			"x_etrs89":          b.XETRS89,
			"y_etrs89":          b.YETRS89,
			"geometry_etrs89":   b.GeometryETRS89,
			"longitude":         b.Longitude,
			"latitude":          b.Latitude,
			"geometry_wgs84":    b.GeometryWGS84,
			"created_at":        b.CreatedAt,
			"deleted_at":        b.DeletedAt,
		})
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *BenchStore) DeleteAllBenches(ctx context.Context) error {
	_, err := s.rdb.Del(ctx, benchesKey).Result()
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

func (s *BenchStore) GetBenchByID(ctx context.Context, gisID string) (*bench.Bench, error) {
	benchKey := fmt.Sprintf("bench:%s", gisID)
	data, err := s.rdb.HGetAll(ctx, benchKey).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	bench := &bench.Bench{
		GisID:            gisID,
		Type:             data["type"],
		Code:             data["code"],
		Description:      data["description"],
		Manufacturer:     data["manufacturer"],
		DistrictCode:     data["district_code"],
		DistrictName:     data["district_name"],
		NeighborhoodCode: data["neighborhood_code"],
		NeighborhoodName: data["neighborhood_name"],
		Zone:             data["zone"],
		StreetName:       data["street_name"],
		StreetNumber:     data["street_number"],
		XETRS89:          data["x_etrs89"],
		YETRS89:          data["y_etrs89"],
		GeometryETRS89:   data["geometry_etrs89"],
		GeometryWGS84:    data["geometry_wgs84"],
	}

	if lat, err := strconv.ParseFloat(data["latitude"], 64); err == nil {
		bench.Latitude = lat
	}
	if lon, err := strconv.ParseFloat(data["longitude"], 64); err == nil {
		bench.Longitude = lon
	}

	return bench, nil
}
