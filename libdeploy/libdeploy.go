package libdeploy

import (
	"strconv"
	"strings"
	"time"
)

const TIME_FORMAT = "2006-01-02T15:04:05Z"

func ParseSetArgument(path string) (string, interface{}) {
	buf := strings.SplitN(path, ":", 2)
	path = buf[0]
	val := buf[1]

	if t, err := time.Parse(TIME_FORMAT, val); err == nil {
		return path, t // Converted to time.Time
	}

	if i, err := strconv.Atoi(val); err == nil {
		return path, i // Converted to int
	}

	if r, err := strconv.ParseFloat(val, 64); err == nil {
		return path, r // Converted to float64
	}

	if b, err := strconv.ParseBool(val); err == nil {
		return path, b // Converted to bool
	}

	return path, val // Cannot conver to any type, sujesting string
}
