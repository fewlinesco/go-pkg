package metrics

import (
	"context"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

// RecordError do a `1` measure each time it receives an error. If error is nil, it's a noop
func RecordError(ctx context.Context, measure *Float64Measure, err error) {
	if err != nil {
		Record(ctx, measure.Measure(1))
	}
}

// RecordElapsedTimeInMilliseconds wraps a callback function and do a measurement of the execution time of the callback
func RecordElapsedTimeInMilliseconds(ctx context.Context, measure *Float64Measure, cb func()) {
	start := time.Now()

	cb()

	Record(ctx, measure.Measure(float64(time.Since(start).Milliseconds())))
}

// Record saves a specific measurement to the current context
func Record(ctx context.Context, m Measurement) {
	stats.Record(ctx, m.Measurement)
}

// RecordWithTags saves a specific measurement to the current context alongside tags
func RecordWithTags(ctx context.Context, tags []Tag, m Measurement) error {
	mutators := make([]tag.Mutator, len(tags))
	for i, t := range tags {
		mutators[i] = tag.Insert(tag.Key(t.Key), t.Value)
	}

	return stats.RecordWithTags(ctx, mutators, m.Measurement)
}
