package utils

import (
	"testing"
	"time"
)

func TestGetWeekDayStr(t *testing.T) {
	type args struct {
		timestamp int64
		day       int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		// 1617225600 is "Thu Apr  1 05:20:00 CST 2021"
		{"TestMon", args{1617225680, 1}, "20210329", false},
		{"TestTue", args{1617225680, 2}, "20210330", false},
		{"TestWed", args{1617225680, 3}, "20210331", false},
		{"TestThu", args{1617225680, 4}, "20210401", false},
		{"TestFri", args{1617225680, 5}, "20210402", false},
		{"TestSat", args{1617225680, 6}, "20210403", false},
		{"TestSun", args{1617225680, 7}, "20210404", false},
		{"TestMon2", args{1617225600 + 3*86400, 1}, "20210329", false},
		{"TestTue2", args{1617225600 + 3*86400, 2}, "20210330", false},
		{"TestWed2", args{1617225600 + 3*86400, 3}, "20210331", false},
		{"TestThu2", args{1617225600 + 3*86400, 4}, "20210401", false},
		{"TestFri2", args{1617225600 + 3*86400, 5}, "20210402", false},
		{"TestSat2", args{1617225600 + 3*86400, 6}, "20210403", false},
		{"TestSun2", args{1617225600 + 3*86400, 7}, "20210404", false},
		{"TestInvalidDay", args{1617225600, 8}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWeekDayStartTime(tt.args.timestamp, tt.args.day)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWeekDayStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			x, _ := time.Parse("20060102", tt.want)
			if got.Unix() != x.Unix() {
				t.Errorf("GetWeekDayStr() got = %v, want %v", got, tt.want)
			}
		})
	}
}
