package collector

// Collector is the interface that abstracts the collection of each metrics.
type Collector interface {
	Register()
	// Collect initiates the collection of metrics.
	Collect() error
	// Destroy cleans up the collector.
	Destroy()
}
