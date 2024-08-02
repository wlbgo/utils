package utils

import (
	"fmt"
	"time"
)

// GetWeekDayStartTime returns the timestamp of the specified day of the week for the given timestamp
// day: 1 for Monday, 2 for Tuesday, ..., 7 for Sunday
func GetWeekDayStartTime(timestamp int64, day int) (time.Time, error) {
	if day < 1 || day > 7 {
		return time.Now(), fmt.Errorf("day must be between 1 and 7")
	}

	// 将时间戳转换为 time.Time 对象
	t := time.Unix(timestamp, 0)
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// 获取当前日期的星期几（0 表示周日，1 表示周一，...，6 表示周六）
	weekday := t.Weekday()

	// 找到上周日
	var lastSun time.Time
	if weekday == 0 {
		lastSun = t.AddDate(0, 0, -7)
	} else {
		lastSun = t.AddDate(0, 0, -int(weekday))
	}
	return lastSun.AddDate(0, 0, day), nil
}
