package metrics

// MetricsEngine is a generic interface to record PBS metrics into the desired backend
type MetricsEngine interface {
	RecordOpenWrapServerPanicStats()
}
