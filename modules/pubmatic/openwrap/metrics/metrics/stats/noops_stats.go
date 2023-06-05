package stats

type noStats struct{}

func (ns *noStats) RecordOpenWrapServerPanicStats() {}
