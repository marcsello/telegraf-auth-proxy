package handlers

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockMetric struct {
	telegraf.Metric
	GetTagFn func(string) (string, bool)
}

func (m MockMetric) GetTag(key string) (string, bool) {
	return m.GetTagFn(key)
}

func TestValidateMetrics(t *testing.T) {
	testcases := []struct {
		name          string
		tagToCheck    string
		expectedValue string
		metrics       []telegraf.Metric
		expectedErr   error
	}{
		{
			name:          "simple_valid",
			tagToCheck:    "host",
			expectedValue: "my-computer",
			metrics: []telegraf.Metric{
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "my-computer", true
					},
				},
			},
			expectedErr: nil,
		}, {
			name:          "multi_valid",
			tagToCheck:    "host",
			expectedValue: "my-computer",
			metrics: []telegraf.Metric{
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "my-computer", true
					},
				},
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "my-computer", true
					},
				},
			},
			expectedErr: nil,
		}, {
			name:          "simple_not_set",
			tagToCheck:    "host",
			expectedValue: "my-computer",
			metrics: []telegraf.Metric{
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "", false
					},
				},
			},
			expectedErr: fmt.Errorf("expected tag %s is not present in metric element %d", "host", 0),
		}, {
			name:          "multi_not_set",
			tagToCheck:    "host",
			expectedValue: "my-computer",
			metrics: []telegraf.Metric{
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "", false
					},
				},
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "", false
					},
				},
			},
			expectedErr: fmt.Errorf("expected tag %s is not present in metric element %d", "host", 0),
		}, {
			name:          "some_not_set",
			tagToCheck:    "host",
			expectedValue: "my-computer",
			metrics: []telegraf.Metric{
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "my-computer", true
					},
				},
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "", false
					},
				},
			},
			expectedErr: fmt.Errorf("expected tag %s is not present in metric element %d", "host", 1),
		}, {
			name:          "simple_invalid_set",
			tagToCheck:    "host",
			expectedValue: "my-computer",
			metrics: []telegraf.Metric{
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "our-computer", true
					},
				},
			},
			expectedErr: fmt.Errorf("tag %s have unexpected value %s (expected %s) in metric element %d", "host", "our-computer", "my-computer", 0),
		}, {
			name:          "multi_invalid_set",
			tagToCheck:    "host",
			expectedValue: "my-computer",
			metrics: []telegraf.Metric{
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "our-computer", true
					},
				},
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "our-computer", true
					},
				},
			},
			expectedErr: fmt.Errorf("tag %s have unexpected value %s (expected %s) in metric element %d", "host", "our-computer", "my-computer", 0),
		}, {
			name:          "some_invalid_set",
			tagToCheck:    "host",
			expectedValue: "my-computer",
			metrics: []telegraf.Metric{
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "my-computer", true
					},
				},
				MockMetric{
					GetTagFn: func(key string) (string, bool) {
						assert.Equal(t, "host", key)
						return "our-computer", true
					},
				},
			},
			expectedErr: fmt.Errorf("tag %s have unexpected value %s (expected %s) in metric element %d", "host", "our-computer", "my-computer", 1),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMetrics(tc.tagToCheck, tc.expectedValue, tc.metrics)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
