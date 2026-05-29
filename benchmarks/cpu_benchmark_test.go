package benchmarks

import "testing"

// Unified benchmark suite: CPU + memory in one run.
// Run with:
//
//	go test -bench=BenchmarkAllLoggers -benchmem -run='^$'

func BenchmarkAllLoggersSimple(b *testing.B) {
	runScenario(b, scenarioSimple, true)
}

func BenchmarkAllLoggersWithFields(b *testing.B) {
	runScenario(b, scenarioWithFields, true)
}

func BenchmarkAllLoggersWithLargeFields(b *testing.B) {
	runScenario(b, scenarioWithLargeFields, true)
}

func BenchmarkAllLoggersWithExtraLargeFields(b *testing.B) {
	runScenario(b, scenarioWithExtraLargeFields, true)
}
