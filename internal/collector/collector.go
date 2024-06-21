package collector

type Metric map[string]interface{}

type MetricContainer []Metric

// Collector is the interface that abstracts the collection of each metrics.
type Collector interface {
	// Register registers the collector.
	Register()
	// Collect initiates the collection of metrics.
	Collect() error
	// PostProcess performs any post-processing of raw data before flushing metrics
	postProcess(metrics MetricContainer) error
}
