package utils

import (
	"fmt"
	"time"
)

// GetWeekDayStr returns the date of the specified day of the week for the given timestamp
// day: 1 for Monday, 2 for Tuesday, ..., 7 for Sunday
func GetWeekDayStr(timestamp int64, day int) (string, error) {
	if day < 1 || day > 7 {
		return "", fmt.Errorf("day must be between 1 and 7")
	}

	// 将时间戳转换为 time.Time 对象
	t := time.Unix(timestamp, 0)

	// 获取当前日期的星期几（0 表示周日，1 表示周一，...，6 表示周六）
	weekday := t.Weekday()

	var weekStart time.Time
	if weekday == 0 {
		weekStart = t.AddDate(0, 0, -7)
	} else {
		weekStart = t.AddDate(0, 0, -int(weekday))
	}

	// 格式化为 %Y%m%d 的字符串
	return weekStart.AddDate(0, 0, day).Format("20060102"), nil
}
