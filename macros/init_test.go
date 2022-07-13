package macros

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
)

// processors for benchmarking add templates
var processor, _ = NewProcessor(STRING_INDEX_CACHED, Config{
	Templates:   []string{""},
	delimiter:   "##",
	valueConfig: MacroValueConfig{},
})

// build sample templates / urls
var totalAccounts = 10000
var eventUrlsPerAccount = 20
var noOfMacrosPerTemplate = 50
var delimiter = "##"
var templatesCount = totalAccounts * eventUrlsPerAccount
var templates = make([]string, 0)

func init() {
	max := 0
	for i := 1; i <= templatesCount; i++ {
		url0 := buildLongInputURL0(noOfMacrosPerTemplate, delimiter) + strconv.Itoa(i)
		if len(url0) > max {
			max = len(url0)
		}
		templates = append(templates, url0)
	}
	fmt.Println(max)
}

func BenchmarkStringCachedIndexBasedAddTemplates(b *testing.B) {
	for n := 0; n < b.N; n++ {
		processor.AddTemplates(templates...)
	}
}
func TestStringIndexCachedBasedAddTemplates(t *testing.T) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	processor.AddTemplates(templates...)
	runtime.ReadMemStats(&m2)
	total := m2.TotalAlloc - m1.TotalAlloc
	malloc := m2.Mallocs - m1.Mallocs
	fmt.Printf("total: %v bytes (%v MB)\n", total, bToMb(total))
	fmt.Printf("mallocs: %v bytes (%v MB)\n", malloc, bToMb(malloc))
	t.Fail()
}

// // PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// // of garage collection cycles completed.
// func PrintMemUsage() {
// 	var m runtime.MemStats
// 	runtime.ReadMemStats(&m)
// 	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
// 	fmt.Println()
// 	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
// 	fmt.Printf("\nTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
// 	fmt.Printf("\nSys = %v MiB", bToMb(m.Sys))
// 	fmt.Printf("\nNumGC = %v\n", m.NumGC)
// }

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
