package tools

import (
	"strings"
	"time"
)

// ParseDate get local time of a date
func ParseDate(thedate string) (time.Time, error) {
	local := time.Now()
	cleanDate := strings.TrimSpace(strings.Replace(thedate, "  ", " ", -1))

	if strings.Index(cleanDate, ":") >= 0 {
		// t, err := time.ParseInLocation("2006-01-02 15:04:05", cleanDate, local.Location())
		t, err := time.ParseInLocation("2006 Jan 2 15:04:05", cleanDate, local.Location())
		
		if err != nil {
			if err != nil {
				return time.Time{}, err
			}
		}
		
		return t, nil
	} else {
		t, err := time.ParseInLocation("2006 Jan 2", cleanDate, local.Location())
		if err != nil {
			if err != nil {
				return time.Time{}, err
			}
		}
		return t, nil
	}
}
