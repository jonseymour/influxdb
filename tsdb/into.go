package tsdb

import (
	"errors"
	"fmt"
	"time"

	"github.com/influxdb/influxdb/influxql"
	"github.com/influxdb/influxdb/models"
)

// EnsureUniqueNonEmptyNames returns an array of len(in) such that out[i] = in[i] iff
// in[i] is unique in in and non-blank. Disambiguation of duplicate and empty names is
// achieved by append _{i} to in[i] where {i} is the string representation of the
// column index i.
func EnsureUniqueNonEmptyNames(in []string) []string {
	out := make([]string, len(in))
	seen := map[string]bool{}
	for i, c := range in {
		n := c
		if c == "" {
			n = fmt.Sprintf("_%d", i)
		}

		for seen[n] { // note: boolean test, not range
			n = fmt.Sprintf("%s_%d", n, i)
		}
		out[i] = n
		seen[n] = true

	}
	return out
}

// convertRowToPoints will convert a query result Row into Points that can be written back in.
// Used for INTO queries
func convertRowToPoints(measurementName string, row *models.Row) ([]models.Point, error) {
	// figure out which parts of the result are the time and which are the fields
	timeIndex := -1
	fieldIndexes := make(map[string]int)
	for i, c := range EnsureUniqueNonEmptyNames(row.Columns) {
		if c == "time" {
			timeIndex = i
		} else {
			fieldIndexes[c] = i
		}
	}

	if timeIndex == -1 {
		return nil, errors.New("error finding time index in result")
	}

	points := make([]models.Point, 0, len(row.Values))
	for _, v := range row.Values {
		vals := make(map[string]interface{})
		for fieldName, fieldIndex := range fieldIndexes {
			val := v[fieldIndex]
			if val != nil {
				vals[fieldName] = v[fieldIndex]
			}
		}

		p, err := models.NewPoint(measurementName, row.Tags, vals, v[timeIndex].(time.Time))
		if err != nil {
			// Drop points that can't be stored
			continue
		}

		points = append(points, p)
	}

	return points, nil
}

func intoDB(stmt *influxql.SelectStatement) (string, error) {
	if stmt.Target.Measurement.Database != "" {
		return stmt.Target.Measurement.Database, nil
	}
	return "", errNoDatabaseInTarget
}

var errNoDatabaseInTarget = errors.New("no database in target")

func intoRP(stmt *influxql.SelectStatement) string          { return stmt.Target.Measurement.RetentionPolicy }
func intoMeasurement(stmt *influxql.SelectStatement) string { return stmt.Target.Measurement.Name }
