package shared

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

func ParseRelativeTime(s string) (time.Time, error) {
	if !strings.HasPrefix(s, "-") {
		return time.Time{}, fmt.Errorf("relative time must start with -")
	}

	s = strings.TrimPrefix(s, "-")

	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return time.Time{}, err
		}
		return time.Now().Add(-time.Duration(days) * 24 * time.Hour), nil
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return time.Time{}, err
	}

	return time.Now().Add(-dur), nil
}

func AggregateMetrics(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

func CalculateZScore(value float64, values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	mean := AggregateMetrics(values)

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0
	}

	return (value - mean) / stdDev
}

func FormatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
