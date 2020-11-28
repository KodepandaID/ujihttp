package examples

import (
	"testing"

	"github.com/KodepandaID/ujihttp/pkg/benchmark"
)

func TestGinBenchmarkGET(t *testing.T) {
	r := benchmark.New()
	r.
		Duration(10).
		Concurrent(500).
		Pipeline(1).
		GET("http://localhost:3000").
		Run()
}
