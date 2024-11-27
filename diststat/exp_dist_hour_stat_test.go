package diststat

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func GetRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "test",
	})
}

func TestStatHelper_Init(t *testing.T) {
	rds := GetRedis()
	defer rds.Close()

	sh := &StatHelper{
		Rds:           rds,
		StatKeyPrefix: "test",
		Period:        time.Hour,
		PeriodStart:   time.Now().Add(-time.Hour),
		StateKeyTTL:   time.Hour,
	}

	err := sh.Init()
	assert.NoError(t, err)
	assert.True(t, sh.inited)
	assert.NotNil(t, sh.inChan)
	assert.NotNil(t, sh.counter)
	assert.NotEmpty(t, sh.workerUUID)
}

func TestStatHelper_CounterIncr(t *testing.T) {
	rds := GetRedis()
	defer rds.Close()

	sh := &StatHelper{
		Rds:           rds,
		StatKeyPrefix: "test",
		Period:        time.Hour,
		PeriodStart:   time.Now().Add(-time.Hour),
		StateKeyTTL:   time.Hour,
	}

	err := sh.Init()
	assert.NoError(t, err)

	sh.CounterIncr("test_label")
	time.Sleep(100 * time.Millisecond) // 等待 worker 处理

	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	assert.Equal(t, 1, sh.counter["test_label"])
}

func TestStatHelper_calcStartPeriod(t *testing.T) {
	sh := &StatHelper{
		Period:      time.Hour,
		PeriodStart: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	t1 := time.Date(2023, 1, 1, 1, 30, 0, 0, time.UTC)
	expected := time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC)
	actual := sh.calcStartPeriod(t1)
	assert.Equal(t, expected, actual)
}

func TestStatHelper_updateCounter(t *testing.T) {
	rds := GetRedis()
	defer rds.Close()

	sh := &StatHelper{
		Rds:           rds,
		StatKeyPrefix: "test",
		Period:        time.Hour,
		PeriodStart:   time.Now().Add(-time.Hour),
		StateKeyTTL:   time.Hour,
	}

	err := sh.Init()
	assert.NoError(t, err)

	sh.CounterIncr("test_label")
	time.Sleep(100 * time.Millisecond) // 等待 worker 处理

	sh.updateCounter()

	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	assert.Equal(t, 1, sh.counter["test_label"])
}

func TestStatHelper_flushCounter(t *testing.T) {
	rds := GetRedis()
	defer rds.Close()

	sh := &StatHelper{
		Rds:           rds,
		StatKeyPrefix: "test",
		Period:        time.Hour,
		PeriodStart:   time.Now().Add(-time.Hour).Truncate(time.Hour),
		StateKeyTTL:   time.Hour,
		FlushPeriod:   time.Second,
	}

	fetchLabelVal := func() (string, error) {
		ctx := context.Background()
		relKey := sh.StatKeyPrefix + ":" + sh.currStartPeriodStart.Format("20060102150405")
		return rds.HGet(ctx, relKey, "test_label:"+sh.workerUUID).Result()
	}

	err := sh.Init()
	assert.NoError(t, err)

	sh.CounterIncr("test_label")
	time.Sleep(100 * time.Millisecond) // 等待 worker 处理

	sh.flushCounter()
	time.Sleep(100 * time.Millisecond) // 等待 worker 处理
	val, err := fetchLabelVal()
	assert.NoError(t, err)
	assert.Equal(t, "1", val)

	sh.CounterIncr("test_label")
	time.Sleep(1100 * time.Millisecond) // 等待 worker 处理

	val, err = fetchLabelVal()
	assert.NoError(t, err)
	assert.Equal(t, "2", val)
}
