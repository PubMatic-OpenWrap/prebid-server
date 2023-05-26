package stats

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

type statLoggerTest struct{}

func (l statLoggerTest) Error(format string, args ...interface{}) {
	//fmt.Errorf(format, args...)
}

func (l statLoggerTest) Info(format string, args ...interface{}) {
	//fmt.Printf(format, args...)
}

func TestInvalidConfig(t *testing.T) {

	l := statLoggerTest{}

	cfg := Config{
		Host:                "",
		Port:                "8090",
		Server:              "SERVER",
		DC:                  "DC",
		PublishingInterval:  2,
		Retries:             1,
		DialTimeout:         3,
		KeepAliveDuration:   30,
		MaxIdleConns:        1,
		MaxIdleConnsPerHost: 2,
		retryInterval:       2,
	}

	_, err := NewClient(cfg, l)
	if err == nil {
		t.Errorf("Should get error")
	}

}

func TestPublishStat(t *testing.T) {

	l := statLogger{}

	cfg := Config{
		Host:                "192.168.0.1",
		Port:                "8090",
		Server:              "SERVER",
		DC:                  "DC",
		PublishingInterval:  2,
		Retries:             1,
		DialTimeout:         3,
		KeepAliveDuration:   30,
		MaxIdleConns:        1,
		MaxIdleConnsPerHost: 2,
		retryInterval:       2,
	}

	client, err := NewClient(cfg, l)
	if err != nil {
		t.Errorf("error creating stats client: %v", err)
	}

	client.PublishStat("TestKey_0", 10)
	client.PublishStat("TestKey_1", 10)
	for i := 0; i < 11000; i++ {
		key := fmt.Sprintf("TestKey_%d", i)
		client.PublishStat(key, 10)
	}

}

var (
	testData = []string{"abcd", "10"}
)

func CreateKeyString(keys ...string) string {
	var b strings.Builder
	for i := 0; i < len(keys); i++ {
		b.WriteString(keys[i])
	}
	return b.String()
}

func BenchmarkCreateKeyString2(b *testing.B) {
	var result string
	for n := 0; n < b.N; n++ {
		result = CreateKeyString("a", "b")
	}
	_ = result
}

func BenchmarkJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := strings.Join(testData, ":")
		_ = s
	}
}

func BenchmarkSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := fmt.Sprintf("%s:%s", testData[0], testData[1])
		_ = s
	}
}

func BenchmarkBufferWithReset(b *testing.B) {
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		buf.Reset()

		buf.WriteString(testData[0])
		buf.WriteString(testData[1])
		s := buf.String()
		_ = s
	}
}
