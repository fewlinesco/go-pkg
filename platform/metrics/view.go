package metrics

import (
	"fmt"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// ViewAggregation represents how to aggregate data before displaying them to the backend
type ViewAggregation struct {
	*view.Aggregation
}

// ViewDistribution organizes a view where measurement will be segmented using the bounds.
// Be careful each bound create a new dimension for the metrics, so if the distribution is used
// with tags it means we can a product of each bound * each tag which can quickly become a huge amount
// of data.
func ViewDistribution(bounds ...float64) *ViewAggregation {
	return &ViewAggregation{view.Distribution(bounds...)}
}

// ViewCount organizes a view where the measurement will be simply sumed.
func ViewCount() *ViewAggregation {
	return &ViewAggregation{view.Count()}
}

// measurer reprensents how we get a measurement from the underlying opencensus library
// It's used internally to cast our structs to opencsensus structs
type measurer interface {
	getStatsMeasure() stats.Measure
}

// View represents a metric to display the backend. It ties a measure an aggregation and a filtering based on scope.
// A view can generate a lot of metrics because when using tag, each tag value because one more dimension for the metric.
type View struct {
	Name        string
	Measure     measurer
	Description string
	TagKeys     []TagKey
	Aggregation *ViewAggregation
}

// RegisterViews is the function to call to register new metric views for display on the HTTP handler.
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
