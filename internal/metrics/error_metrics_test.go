package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorMetrics(t *testing.T) {
	metrics := NewErrorMetrics()

	t.Run("Record and Retrieve Error", func(t *testing.T) {
		metrics.Record("auth_failed", 401, 0.5, "/login")
		metrics.Record("auth_failed", 401, 0.7, "/login")
		metrics.Record("not_found", 404, 0.3, "/unknown")

		stats := metrics.GetStats()
		errorCounts := stats["error_counts"].(map[string]*ErrorEntry)

		assert.Equal(t, int64(2), errorCounts["auth_failed"].Count)
		assert.Equal(t, int64(1), errorCounts["not_found"].Count)
		assert.InDelta(t, 0.6, errorCounts["auth_failed"].AvgLatency(), 0.01)
	})

	t.Run("Get Recent Errors", func(t *testing.T) {
		metrics := NewErrorMetrics()
		metrics.Record("error1", 500, 0.1, "/path1")
		time.Sleep(time.Millisecond)
		metrics.Record("error2", 400, 0.2, "/path2")

		recent := metrics.GetRecentErrors(2)
		assert.Equal(t, 2, len(recent))
		assert.Equal(t, "error2", recent[0].ErrorType)
		assert.Equal(t, "error1", recent[1].ErrorType)
	})

	t.Run("Get Error Rates", func(t *testing.T) {
		metrics := NewErrorMetrics()

		// Use a fixed reference time
		refTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

		// Add some errors at different times - make them very distinct
		twoHoursAgo := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)      // 2 hours before refTime
		fiftyMinutesAgo := time.Date(2025, 1, 1, 11, 10, 0, 0, time.UTC) // 50 minutes before refTime
		justNow := refTime

		t.Logf("Reference time: %v", refTime)
		t.Logf("50 minutes ago: %v", fiftyMinutesAgo)
		t.Logf("Two hours ago: %v", twoHoursAgo)
		t.Logf("Cutoff (1 hour before ref): %v", time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC))

		// Record the errors
		metrics.recordWithTime("rate_error", 500, 0.1, "/path", twoHoursAgo)
		metrics.recordWithTime("rate_error", 500, 0.1, "/path", fiftyMinutesAgo)
		metrics.recordWithTime("rate_error", 500, 0.1, "/path", justNow)

		// Print timestamps stored in the metrics
		entry := metrics.errors["rate_error"]
		t.Logf("Stored timestamps: %v", entry.Timestamps)

		// Use the same reference time for consistency
		rates := metrics.GetErrorRatesWithReference(time.Hour, refTime)
		t.Logf("Rates: %v", rates)

		// We expect 2 errors in the last hour: the one from fiftyMinutesAgo and the one from justNow
		assert.Equal(t, float64(2), rates["rate_error"])
	})

	t.Run("Get Latency Percentiles", func(t *testing.T) {
		metrics := NewErrorMetrics()

		// Add errors with various latencies
		for i := 0; i < 100; i++ {
			metrics.Record("perf_test", 200, float64(i), "/path")
		}

		percentiles := metrics.GetLatencyPercentiles()
		assert.InDelta(t, 95, percentiles["p95"], 1.0)
		assert.InDelta(t, 50, percentiles["p50"], 1.0)
	})

	t.Run("Cleanup Old Entries", func(t *testing.T) {
		metrics := NewErrorMetrics()
		now := time.Now()

		// Add old and new errors
		metrics.recordWithTime("old_error", 500, 0.1, "/path", now.Add(-25*time.Hour))
		metrics.recordWithTime("new_error", 500, 0.1, "/path", now)

		metrics.Cleanup(24 * time.Hour)

		stats := metrics.GetStats()
		errorCounts := stats["error_counts"].(map[string]*ErrorEntry)

		assert.Equal(t, int64(0), errorCounts["old_error"].Count)
		assert.Equal(t, int64(1), errorCounts["new_error"].Count)
	})
}
