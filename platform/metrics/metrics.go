package metrics

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

// Unit describing the base unit. It follows this naming convention http://unitsofmeasure.org/ucum.html
type Unit string

const (
	// UnitDimensionless reprensents a simple counter with no specific unit
	UnitDimensionless Unit = "1"
	// UnitMilliseconds represents a time elapsed in milliseconds
	UnitMilliseconds Unit = "ms"
)

// Measurement represents a specific measure of a stat
type Measurement struct{ stats.Measurement }

// TagKey represents the name of a tag
type TagKey tag.Key

// Tag represents additional data to use alongside a measure.
// Be careful with tags, the more values possible for a tag, the more dimensions the data will have.
// For example:
// - identifiers are bad tags because it has an opened and wide range of values
// - A response code or a status can make a good tag because the number of possible values is small and well defined
type Tag struct {
	Key   TagKey
	Value string
}

// MustNewTagKey is a constructor to create a new tag key.
// It must only contains basic ASCII characters (space + all visible characters)
func MustNewTagKey(t string) TagKey {
	return TagKey(tag.MustNewKey(t))
}
