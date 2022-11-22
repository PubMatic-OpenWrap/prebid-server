package floors

import "github.com/prebid/prebid-server/config"

type FetchInfo struct {
	config.AccountFloorFetch
	FetchPeriod int64
}

type FetchQueue []*FetchInfo

func (fq FetchQueue) Len() int {
	return len(fq)
}

func (fq FetchQueue) Less(i, j int) bool {
	return fq[i].Period < fq[j].Period
}

func (fq FetchQueue) Swap(i, j int) {
	fq[i], fq[j] = fq[j], fq[i]
}

func (fq *FetchQueue) Push(element interface{}) {
	fetchInfo := element.(*FetchInfo)
	*fq = append(*fq, fetchInfo)
}

func (fq *FetchQueue) Pop() interface{} {
	old := *fq
	n := len(old)
	fetchInfo := old[n-1]
	old[n-1] = nil // avoid memory leak
	*fq = old[0 : n-1]
	return fetchInfo
}

func (fq *FetchQueue) Top() *FetchInfo {
	old := *fq
	if len(old) == 0 {
		return nil
	}
	return old[0]
}
