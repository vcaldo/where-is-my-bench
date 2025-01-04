package bench

import (
	"context"
	"encoding/json"

	"github.com/newrelic/go-agent/v3/newrelic"
)

type Bench struct {
	GisID            string  `json:"gis_id"`
	Type             string  `json:"tipus_de_mobiliari_urba"`
	Code             string  `json:"codi"`
	Description      string  `json:"descripcio"`
	Manufacturer     string  `json:"fabricant"`
	DistrictCode     string  `json:"codi_districte"`
	DistrictName     string  `json:"nom_districte"`
	NeighborhoodCode string  `json:"codi_barri"`
	NeighborhoodName string  `json:"nom_barri"`
	Zone             string  `json:"zona"`
	StreetName       string  `json:"nom_carrer"`
	StreetNumber     string  `json:"num_carrer"`
	XETRS89          string  `json:"x_etrs89"`
	YETRS89          string  `json:"y_etrs89"`
	GeometryETRS89   string  `json:"geometria_etrs89"`
	Longitude        float64 `json:"longitud"`
	Latitude         float64 `json:"latitud"`
	GeometryWGS84    string  `json:"geometria_wgs84"`
	CreatedAt        string  `json:"data_alta"`
	DeletedAt        string  `json:"data_baixa"`
}

func LoadBenches(ctx context.Context, jsonData []byte) ([]Bench, error) {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("load_benches")
	defer segment.End()

	txn.AddAttribute("json_size_bytes", len(jsonData))

	var benches []Bench
	err := json.Unmarshal(jsonData, &benches)
	if err != nil {
		return nil, err
	}

	txn.AddAttribute("benches_count", len(benches))

	return benches, nil
}
