package libdeploy

import (
	"strconv"
	"strings"
	"time"
)

const TIME_FORMAT = "2006-01-02T15:04:05Z"

func ParseSetArgument(path string) (interface{}, string) {
	buf := strings.SplitN(path, ":", 2)
	path = buf[0]
	val := buf[1]

	if t, err := time.Parse(TIME_FORMAT, val); err == nil {
		return t, path // Converted to time.Time
	}

	if i, err := strconv.Atoi(val); err == nil {
		return i, path // Converted to int
	}

	if r, err := strconv.ParseFloat(val, 64); err == nil {
		return r, path // Converted to float64
	}

	if b, err := strconv.ParseBool(val); err == nil {
		return b, path // Converted to bool
	}

	return val, path // Cannot conver to any type, suggesting string
}
