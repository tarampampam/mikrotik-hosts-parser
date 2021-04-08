package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Generator contains script generation metric collectors.
type Generator struct {
	cacheHit  prometheus.Counter
	cacheMiss prometheus.Counter
	duration  prometheus.Histogram
}

// NewGenerator creates new Generator metrics collector.
func NewGenerator() Generator {
	return Generator{
		cacheHit: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "generator",
			Subsystem: "cache",
			Name:      "hits",
			Help:      "The count of cache hits during script generation.",
		}),
		cacheMiss: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "generator",
			Subsystem: "cache",
			Name:      "misses",
			Help:      "The count of cache misses during script generation.",
		}),
		duration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "generator",
			Subsystem: "time",
			Name:      "duration",
			Help:      "Time of script generation (in seconds).",
		}),
	}
}

// IncrementCacheHits increments cache hits counter.
func (g *Generator) IncrementCacheHits() { g.cacheHit.Inc() }

// IncrementCacheMisses increments cache misses counter.
func (g *Generator) IncrementCacheMisses() { g.cacheMiss.Inc() }

// ObserveGenerationDuration adds a single observation to the script generation histogram.
func (g *Generator) ObserveGenerationDuration(d time.Duration) { g.duration.Observe(d.Seconds()) }

// Register metrics with registerer.
func (g *Generator) Register(reg prometheus.Registerer) error {
	for _, c := range [...]prometheus.Collector{g.cacheHit, g.cacheMiss, g.duration} {
		if e := reg.Register(c); e != nil {
			return e
		}
	}

	return nil
}
