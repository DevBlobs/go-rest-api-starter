package tz

import (
	"fmt"
	"time"
)

const CompanyTimezone = "Asia/Tokyo"

var location *time.Location

func init() {
	var err error
	location, err = time.LoadLocation(CompanyTimezone)
	if err != nil {
		panic(fmt.Sprintf("failed to load timezone %s: %v", CompanyTimezone, err))
	}
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("20060102", s, location)
}
