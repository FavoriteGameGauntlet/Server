package dbaccess

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ScanInterval parses a Postgres INTERVAL value as decoded by the pgx stdlib
// driver, which returns it as text in "[-]HHH:MM:SS[.ffffff]" form for
// values built purely from hours/minutes/seconds (no day/month components).
func ScanInterval(raw string) (time.Duration, error) {
	negative := strings.HasPrefix(raw, "-")
	if negative {
		raw = raw[1:]
	}

	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("dbaccess: unexpected interval format %q", raw)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("dbaccess: unexpected interval format %q: %w", raw, err)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("dbaccess: unexpected interval format %q: %w", raw, err)
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, fmt.Errorf("dbaccess: unexpected interval format %q: %w", raw, err)
	}

	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds*float64(time.Second))

	if negative {
		duration = -duration
	}

	return duration, nil
}
