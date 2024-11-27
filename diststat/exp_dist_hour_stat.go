package diststat

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"strconv"
	"sync"
	"time"
)

type StatHelper struct {
	Rds           *redis.Client
	StatKeyPrefix string
	Period        time.Duration
	PeriodStart   time.Time
	StateKeyTTL   time.Duration
	FlushPeriod   time.Duration // default 1s

	// hide the details
	inited                bool
	currStartPeriodStart  time.Time
	workerUUID            string
	inChan                chan string
	counter               map[string]int
	updatedSinceLastFlush bool
	mutex                 sync.Mutex
}

func (s *StatHelper) Init() error {
	if s.inited {
		return errors.New("multiple init")
	}

	// check values
	if s.Rds == nil {
		return errors.New("redis is nil")
	}
	if s.StatKeyPrefix == "" {
		return errors.New("StatKeyPrefix is empty")
	}

	if s.Period == 0 {
		return errors.New("period is 0")
	}
	if s.PeriodStart.IsZero() {
		return errors.New("PeriodStart is zero")
	}
	if s.Period%time.Second != 0 {
		return errors.New("period is not an integer multiple of one second")
	}
	s.currStartPeriodStart = s.calcStartPeriod(time.Now())
	s.inChan = make(chan string, 10000)
	s.counter = make(map[string]int)
	hash := md5.Sum([]byte(uuid.New().String()))
	s.workerUUID = hex.EncodeToString(hash[:])[:8]
	s.updatedSinceLastFlush = false
	s.inited = true
	go s.worker()
	return nil
}

func (s *StatHelper) CounterIncr(label string) {
	s.inChan <- label
	//s.mutex.Lock()
	//defer s.mutex.Unlock()
	s.updatedSinceLastFlush = true
}

func (s *StatHelper) worker() {
	flushPeriod := time.Second
	if s.FlushPeriod != 0 {
		flushPeriod = s.FlushPeriod
	}
	ticker := time.NewTicker(flushPeriod)
	defer ticker.Stop()
	for {
		select {
		case k := <-s.inChan:
			s.mutex.Lock()
			s.counter[k]++
			s.mutex.Unlock()
		case <-ticker.C:
			s.updateCounter()
		}
	}
}

func (s *StatHelper) calcStartPeriod(t time.Time) time.Time {
	return s.PeriodStart.Add(t.Sub(s.PeriodStart).Truncate(s.Period))
}

func (s *StatHelper) updateCounter() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	now := time.Now()
	if s.currStartPeriodStart.Add(s.Period).Before(now) {
		s.flushCounter()
		s.currStartPeriodStart = s.calcStartPeriod(now)
		s.counter = make(map[string]int)
		return
	}
	s.flushCounter()
}

func (s *StatHelper) flushCounter() {
	if !s.updatedSinceLastFlush {
		return
	}

	s.updatedSinceLastFlush = false
	kvList := make([]string, 0, len(s.counter)*2)
	relKey := s.StatKeyPrefix + ":" + s.currStartPeriodStart.Format("20060102150405")
	for key, val := range s.counter {
		subKey := key + ":" + s.workerUUID
		kvList = append(kvList, subKey, strconv.Itoa(val))
	}

	if len(kvList) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Rds.HSet(ctx, relKey, kvList).Err()
	if err != nil {
		err := s.Rds.HSet(ctx, relKey, kvList).Err()
		if err != nil {
			// TODO
		}
	}

	if s.StateKeyTTL != 0 {
		s.Rds.Expire(ctx, relKey, s.StateKeyTTL)
	}
}
