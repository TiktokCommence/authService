package tool

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

func ParseTimeString(s string) (time.Duration, error) {
	re, err := regexp.Compile(`(\d+)([hmns]+)`)
	if err != nil {
		return 0, err
	}
	matches := re.FindStringSubmatch(s)
	numstr := matches[1]
	unit := matches[2]

	num, err := strconv.ParseInt(numstr, 10, 64)
	if err != nil {
		return 0, err
	}
	var td time.Duration
	switch unit {
	case "ns":
		td = time.Duration(num) * time.Nanosecond
	case "ms":
		td = time.Duration(num) * time.Millisecond
	case "s":
		td = time.Duration(num) * time.Second
	case "m":
		td = time.Duration(num) * time.Minute
	case "h":
		td = time.Duration(num) * time.Hour
	default:
		return 0, errors.New("invalid unit")
	}
	return td, nil
}
