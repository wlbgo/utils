package utils

import (
	"reflect"
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
		{"TestMon1", args{1722502012, 1}, "20240729", false},
		{"TestMon2", args{1722583522, 1}, "20240729", false},
		{"TestInvalidDay", args{1617225600, 8}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWeekDayStartTime(tt.args.timestamp, tt.args.day)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWeekDayStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				x, _ := time.ParseInLocation("20060102", tt.want, time.Local)
				if got.Unix() != x.Unix() {
					t.Errorf("GetWeekDayStr() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestNearlyPeriodStartTime(t *testing.T) {
	type args struct {
		start  time.Time
		period time.Duration
		ts     time.Time
	}
	t1, _ := time.ParseInLocation("2006-01-02 15:04:05", "2021-08-13 00:00:00", time.Local)
	t2, _ := time.ParseInLocation("2006-01-02 15:04:05", "2021-08-13 01:00:00", time.Local)
	t3, _ := time.ParseInLocation("2006-01-02 15:04:05", "2021-08-13 00:00:00", time.Local)
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{name: "Test1", args: args{
			start:  t1,
			period: time.Hour * 24,
			ts:     t2,
		},
			want: t3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NearlyPeriodStartTime(tt.args.start, tt.args.period, tt.args.ts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nearlyPeriodStartTime() = %v, want %v", got, tt.want)
			}
		})
	}

}
