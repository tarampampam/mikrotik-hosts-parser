package metrics_test

import (
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/metrics"
)

func TestGenerator_Register(t *testing.T) {
	var (
		registry = prometheus.NewRegistry()
		gen      = metrics.NewGenerator()
	)

	assert.NoError(t, gen.Register(registry))

	count, err := testutil.GatherAndCount(registry,
		"generator_cache_hits",
		"generator_cache_misses",
		"generator_time_duration",
	)
	assert.NoError(t, err)

	assert.Equal(t, 3, count)
}

func TestGenerator_IncrementCacheHits(t *testing.T) {
	gen := metrics.NewGenerator()

	gen.IncrementCacheHits()

	metric := getMetric(&gen, "generator_cache_hits")
	assert.Equal(t, float64(1), metric.Counter.GetValue())
}

func TestGenerator_IncrementCacheMisses(t *testing.T) {
	gen := metrics.NewGenerator()

	gen.IncrementCacheMisses()

	metric := getMetric(&gen, "generator_cache_misses")
	assert.Equal(t, float64(1), metric.Counter.GetValue())
}

func TestGenerator_ObserveGenerationDuration(t *testing.T) {
	gen := metrics.NewGenerator()

	gen.ObserveGenerationDuration(time.Second)
	gen.ObserveGenerationDuration(2 * time.Second)

	metric := getMetric(&gen, "generator_time_duration")
	assert.Equal(t, float64(3), metric.Histogram.GetSampleSum())
	assert.Equal(t, uint64(2), metric.Histogram.GetSampleCount())
}

type registerer interface {
	Register(prometheus.Registerer) error
}

func getMetric(m registerer, name string) *dto.Metric {
	registry := prometheus.NewRegistry()
	_ = m.Register(registry)

	families, _ := registry.Gather()

	for _, family := range families {
		if family.GetName() == name {
			return family.Metric[0]
		}
	}

	return nil
}
