//go:build !arm64

package metrics

// NewMetricsEngineMock creates a new instance of MetricsEngineMock
func NewMetricsEngineMock() *MetricsEngineMock {
	return &MetricsEngineMock{}
}
