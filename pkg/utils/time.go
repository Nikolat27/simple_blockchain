package utils

import "time"

// GetTimestamp -> Returns current time in utc unix milliseconds
func GetTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}
