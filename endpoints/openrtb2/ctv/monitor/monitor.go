// Package monitor provides the mechanism for measuring and monitoring
// the function / processes
package monitor

import "time"

/*
IMonitor provides the mechanism to monitor the function/process performance
under consideration using following metrics
	1. Execution time - Elapsed time taken in Nanoseconds. Observations are counted in buckets of 10000 (ns) 40000 (ns), 80000 (ns), 100000 (ns)
*/
type IMonitor interface {
	/*
		MeasureExecutionTime computes the time taken by process/function to complete
		it task. Typically this function is called using `defer` keyword
			Example:
			start := time.Now()
			defer monitor.MeasureExecutionTime(start)
	*/
	MeasureExecutionTime(time.Time)

	/*
		Scenario provides information around what is being measured
	*/
	Scenario(string)
}

/*
New returns the instance of IMonitor object
*/
func New(algorithm string) IMonitor {
	return newPrometheusMonitor(algorithm)
}
