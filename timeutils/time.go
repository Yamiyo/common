package timeutils

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	DESC = "desc"
	ASC  = "asc"

	granularityHour = 60
	granularityDay  = 1440
	hourSevenDays   = 168

	timeFormatDay  = "%Y-%m-%d"
	timeFormatHour = "%Y-%m-%d %H:00:00"

	Format_DateTime_TimeZone_TZ = "2006-01-02T15:04:05Z07:00"
	Format_DateTime_TimeZone    = "2006-01-02 15:04:05-07:00"
	Format_DateTime             = "2006-01-02 15:04:05"
	Format_DateTimeMilliSecond  = "2006-01-02 15:04:05.999"
	Format_DateTimeMicroSecond  = "2006-01-02 15:04:05.999999"
	Format_DateTimeNanoSecond   = "2006-01-02 15:04:05.999999999"
	Format_DateHour             = "2006-01-02 15:00:00"
	Format_DateHourMinute       = "2006-01-02 15:04:00"
	Format_Date                 = "2006-01-02"
	Format_Time                 = "15:04:05"
	Format_YYYYMMDD             = "20060102"
)

type PickTime struct {
	StartTime time.Time
	EndTime   time.Time
}

func ParseTimeRangeToPickTimes(sTime, eTime time.Time, interval int) []*PickTime {
	intervalDuring := time.Duration(interval) * time.Minute
	timeRange := eTime.Sub(sTime)
	pickTimeNo := int(math.Ceil(timeRange.Minutes() / intervalDuring.Minutes()))

	var pickTimes []*PickTime
	for idx := 0; idx < pickTimeNo; idx++ {
		tmpPickTime := PickTime{
			StartTime: sTime.Add(time.Duration(interval*idx) * time.Minute),
			EndTime:   sTime.Add(time.Duration(interval*(idx+1)) * time.Minute),
		}

		if tmpPickTime.EndTime.After(eTime) {
			tmpPickTime.EndTime = eTime
		}
		pickTimes = append(pickTimes, &tmpPickTime)
	}
	return pickTimes
}

func GetMonthSlice(sTime, eTime time.Time) []*PickTime {
	firstDateOfMonth := GetFirstDateOfMonth(eTime)
	if firstDateOfMonth.Equal(eTime) {
		firstDateOfMonth = firstDateOfMonth.AddDate(0, -1, 0)
	}
	var pickTimes []*PickTime
	if firstDateOfMonth.After(sTime) {
		tmpPickTime := PickTime{
			StartTime: firstDateOfMonth,
			EndTime:   eTime,
		}
		pickTimes = append(pickTimes, GetMonthSlice(sTime, firstDateOfMonth)...)
		pickTimes = append(pickTimes, &tmpPickTime)
	} else {
		tmpPickTime := PickTime{
			StartTime: sTime,
			EndTime:   eTime,
		}
		pickTimes = append(pickTimes, &tmpPickTime)
	}

	return pickTimes
}

func GetFirstDateOfMonth(d time.Time) time.Time {
	d = d.AddDate(0, 0, -d.Day()+1)
	return GetZeroTime(d)
}

func GetZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

// GetDateTimeSlice granularity unit = minute.
func GetDateTimeSlice(st, et time.Time, granularity int, sequence string) []string {
	layout := Format_Date
	if granularity == granularityHour {
		layout = Format_DateHour
	}

	ct := st
	var dateTimeSlice []string
	for {
		dateTimeSlice = append(dateTimeSlice, ct.Format(layout))
		ct = ct.Add(time.Minute * time.Duration(granularity))
		if et.Sub(ct).Seconds() <= 0 {
			break
		}
	}

	if sequence == DESC {
		sort.Slice(dateTimeSlice, func(i, j int) bool {
			return dateTimeSlice[i] > dateTimeSlice[j]
		})
	}

	return dateTimeSlice
}

func CloneTimePtr(t *time.Time) (*time.Time, error) {
	clone := time.Time{}

	b, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if err := clone.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return &clone, nil
}

func Time2String(t *time.Time, delim string) string {
	ts := []int{t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second()}
	format := []string{"%d", "%02d", "%02d", "%02d", "%02d", "%02d"}
	tss := make([]string, 6)
	for i, v := range ts {
		tss[i] = fmt.Sprintf(format[i], v)
	}
	return strings.Join(tss, delim)
}

func FixedZone(offset int) *time.Location {
	return time.FixedZone("", offset*60*60)
}

func ConvertDateTimeRngByTimezone(startTime, endTime string, srcOffset int) (*time.Time, *time.Time) {
	sTime, eTime := getDateTimeRngByTimezone(startTime, endTime, srcOffset)

	if sTime == nil || eTime == nil {
		return sTime, eTime
	}

	return fixTimeOrder(sTime, eTime)
}

func getDateTimeRngByTimezone(startTime, endTime string, srcOffset int) (*time.Time, *time.Time) {
	tz := GetTimeZoneString(srcOffset)

	sTime := TryParseTime(PadTime(startTime)+tz, Format_DateTime_TimeZone, nil)
	eTime := TryParseTime(PadTime(endTime)+tz, Format_DateTime_TimeZone, nil)

	return sTime, eTime
}

func GetTimeZoneString(offset int) string {
	if offset < 0 {
		return fmt.Sprintf("%03d:00", offset)
	}

	return fmt.Sprintf("+%02d:00", offset)
}

//Try to parse string with specified layout, if not, return d
func TryParseTime(v string, layout string, d *time.Time) *time.Time {
	if t, err := time.Parse(layout, v); err == nil {
		return &t
	} else {
		return d
	}
}

func PadTime(source string) string {
	items := Split(source, " ", true)
	if len(items) == 0 {
		return " 00:00:00"
	} else if len(items) == 1 {
		return items[0] + " 00:00:00"
	} else {
		t := Split(items[1], ":", true)
		if len(t) == 2 {
			return fmt.Sprintf("%s %02s:%02s:00", items[0], t[0], t[1])
		} else {
			return fmt.Sprintf("%s %02s:%02s:%02s", items[0], t[0], t[1], t[2])
		}
	}
}

func Split(s, sep string, removeEmptyItem bool) []string {
	items := make([]string, 0, 10)
	for _, v := range strings.Split(s, sep) {
		v = strings.TrimSpace(v)
		if removeEmptyItem {
			if HasValue(v) {
				items = append(items, v)
			}
		} else {
			items = append(items, v)
		}
	}

	return items
}

func fixTimeOrder(startTime, endTime *time.Time) (*time.Time, *time.Time) {
	if startTime == nil || endTime == nil {
		return startTime, endTime
	}

	if startTime.UTC().After(endTime.UTC()) {
		startTime, endTime = endTime, startTime
	}

	return startTime, endTime
}

func HasValue(s string) bool {
	return len(strings.Trim(s, " ")) > 0
}

func TruncateTZ(t time.Time, d time.Duration) time.Time {
	_, zone := t.Zone()
	if zone == 0 {
		return t.Truncate(d)
	}

	utcTime, _ := time.Parse(Format_DateTimeNanoSecond, t.Format(Format_DateTimeNanoSecond))
	truncateTime := utcTime.Truncate(d)
	year, month, day := truncateTime.Date()
	hour := truncateTime.Hour()
	minute := truncateTime.Minute()
	second := truncateTime.Second()
	nanosecond := truncateTime.Nanosecond()
	location := t.Location()
	outTime := time.Date(year, month, day, hour, minute, second, nanosecond, location)

	return outTime
}
