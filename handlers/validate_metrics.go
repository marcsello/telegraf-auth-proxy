package handlers

import (
	"fmt"
	"github.com/influxdata/telegraf"
)

func validateMetrics(tagToCheck, expectedValue string, metrics []telegraf.Metric) error {
	// validate tag(s)...
	for i, metric := range metrics {
		fieldValue, ok := metric.GetTag(tagToCheck)
		if !ok {
			// tag was not present, consider invalid
			return fmt.Errorf("expected tag %s is not present in metric element %d", tagToCheck, i)
		}
		if fieldValue != expectedValue {
			// tag have unexpected value, consider invalid
			return fmt.Errorf("tag %s have unexpected value %s (expected %s) in metric element %d", tagToCheck, fieldValue, expectedValue, i)
		}
	}

	// all went well
	return nil
}
