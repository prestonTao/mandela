package utils

import (
	"time"
)

func TimeFormatToNanosecond() string {
	return time.Now().Format("2006-01-02 15:04:05.999999999")
}

func FormatTimeToSecond(now time.Time) string {
	return now.Format("2006-01-02 15:04:05")
}
