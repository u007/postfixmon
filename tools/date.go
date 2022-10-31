package tools

import (
	"strings"
	"time"
)

// ParseDate get local time of a date
func ParseDate(thedate string) (time.Time, error) {
	local := time.Now()
	if strings.Index(thedate, ":") >= 0 {
		// t, err := time.ParseInLocation("2006-01-02 15:04:05", thedate, local.Location())
		t, err := time.ParseInLocation("2006 Jan 2 15:04:05", thedate, local.Location())
		
		if err != nil {
			// double space
			t, err = time.ParseInLocation("2006 Jan  2 15:04:05", thedate, local.Location())
			if err != nil {
				return time.Time{}, err
			}
		}

		
		return t, nil
	} else {
		t, err := time.ParseInLocation("2006 Jan 2", thedate, local.Location())
		if err != nil {
			t, err = time.ParseInLocation("2006 Jan  2", thedate, local.Location())
			if err != nil {
				return time.Time{}, err
			}
		}
		return t, nil
	}
}
