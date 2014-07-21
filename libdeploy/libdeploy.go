package libdeploy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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

func ResolveDomainName(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, ip := range ips {
		result = append( result, ip.String())
	}

	return result, nil
}

func MarshalToJsonReader(value interface{}) (io.Reader, error) {
	buf, err := json.Marshal(value)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error marshaling data: %s", err))
	}

	return bytes.NewBuffer(buf), nil
}
