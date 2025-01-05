package maps

import (
	"context"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"time"

	sm "github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/pkg/bench"
)

type MapGenerator struct {
	ctx *sm.Context
}

func NewMapGenerator() *MapGenerator {
	ctx := sm.NewContext()
	ctx.SetSize(1600, 1200)
	return &MapGenerator{ctx: ctx}
}

func (m *MapGenerator) GenerateMap(ctx context.Context, lat, lon, radius float64, benches []bench.Bench) (string, error) {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("generate_map")
	defer segment.End()

	txn.AddAttribute("latitude", lat)
	txn.AddAttribute("longitude", lon)
	txn.AddAttribute("radius", radius)
	txn.AddAttribute("benches_count", len(benches))

	m.ctx.SetCenter(s2.LatLngFromDegrees(lat, lon))

	circle := sm.NewCircle(s2.LatLngFromDegrees(lat, lon),
		color.RGBA{R: 0, G: 0, B: 0, A: 128},
		color.RGBA{R: 0, G: 0, B: 0, A: 64},
		radius, 4.0)
	m.ctx.AddObject(circle)

	centerMarker := sm.NewMarker(
		s2.LatLngFromDegrees(lat, lon),
		color.RGBA{R: 255, G: 255, B: 0, A: 255},
		24.0,
	)
	m.ctx.AddObject(centerMarker)

	segment = txn.StartSegment("add_benches")
	defer segment.End()
	for _, b := range benches {
		marker := sm.NewMarker(
			s2.LatLngFromDegrees(b.Latitude, b.Longitude),
			color.RGBA{R: 255, G: 0, B: 0, A: 255},
			16.0,
		)
		m.ctx.AddObject(marker)
	}
	segment.End()

	segment = txn.StartSegment("render_map")
	defer segment.End()

	img, err := m.ctx.Render()
	if err != nil {
		return "", err
	}
	segment.End()
	filename := fmt.Sprintf("%d-map_%f_%f.png", time.Now().Unix(), lat, lon)
	f, err := os.Create(filename)
	if err != nil {

		return "", err
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return "", err
	}

	return filename, nil
}
