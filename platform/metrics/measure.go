package metrics

import (
	"go.opencensus.io/stats"
)

// Float64Measure represents one specific float64 stat to measure
type Float64Measure struct{ *stats.Float64Measure }

func (f *Float64Measure) getStatsMeasure() stats.Measure { return f.Float64Measure }

// M create a new measurement for the stat
func (f *Float64Measure) M(v float64) Measurement { return Measurement{f.Float64Measure.M(v)} }

// Float64 is a constructor to define a float64 stat to measure
func Float64(name, description string, unit Unit) *Float64Measure {
	return &Float64Measure{stats.Float64(name, description, string(unit))}
}
