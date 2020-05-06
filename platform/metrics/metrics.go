package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

type Unit string

const (
	UnitDimensionless Unit = "1"
	UnitMilliseconds  Unit = "ms"
)

type TagKey tag.Key

type Tag struct {
	Key   TagKey
	Value string
}

type Measure interface {
	getStatsMeasure() stats.Measure
}

type ViewAggregation struct {
	*view.Aggregation
}

type View struct {
	Name        string
	Measure     Measure
	Description string
	TagKeys     []TagKey
	Aggregation *ViewAggregation
}

type Measurement struct{ stats.Measurement }

type Float64Measure struct{ *stats.Float64Measure }

func (m *Float64Measure) getStatsMeasure() stats.Measure {
	return m.Float64Measure
}

func CreateHandler(namespace string) (http.Handler, error) {
	pe, err := prometheus.NewExporter(prometheus.Options{Namespace: namespace})

	if err != nil {
		return nil, fmt.Errorf("can't create metric exporter: %v", err)
	}

	return pe, nil
}

func RegisterViews(v ...*View) error {
	views := make([]*view.View, len(v))

	for i := range v {
		var tagKeys []tag.Key
		for _, t := range v[i].TagKeys {
			tagKeys = append(tagKeys, tag.Key(t))
		}

		views[i] = &view.View{
			Name:        v[i].Name,
			Measure:     v[i].Measure.getStatsMeasure(),
			Description: v[i].Description,
			TagKeys:     tagKeys,
			Aggregation: v[i].Aggregation.Aggregation,
		}
	}

	if err := view.Register(views...); err != nil {
		return fmt.Errorf("can't register metric views: %v", err)
	}

	return nil
}

func Float64(name, description string, unit Unit) *Float64Measure {
	return &Float64Measure{stats.Float64(name, description, string(unit))}
}

func ViewDistribution(bounds ...float64) *ViewAggregation {
	return &ViewAggregation{view.Distribution(bounds...)}
}

func ViewCount() *ViewAggregation {
	return &ViewAggregation{view.Count()}
}

func RecordError(ctx context.Context, measure *Float64Measure, err error) {
	if err != nil {
		Record(ctx, measure.M(1))
	}
}

func RecordElapsedTimeInMilliseconds(ctx context.Context, measure *Float64Measure, cb func()) {
	start := time.Now()

	cb()

	Record(ctx, measure.M(float64(time.Now().Sub(start).Milliseconds())))
}

func (f *Float64Measure) M(v float64) Measurement {
	return Measurement{f.Float64Measure.M(v)}
}

func MustNewTagKey(t string) TagKey {
	return TagKey(tag.MustNewKey(t))
}

func Record(ctx context.Context, m Measurement) {
	stats.Record(ctx, m.Measurement)
}

func RecordWithTags(ctx context.Context, tags []Tag, m Measurement) error {
	mutators := make([]tag.Mutator, len(tags))
	for i, t := range tags {
		mutators[i] = tag.Insert(tag.Key(t.Key), t.Value)
	}

	return stats.RecordWithTags(ctx, mutators, m.Measurement)
}
