package cachecfg

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestValueFetcher 用于测试的 ValueFetcher 实现
type TestValueFetcher struct {
	counter int64
}

func (f *TestValueFetcher) Key(args ...any) string {
	if len(args) > 0 {
		return fmt.Sprintf("test-key-%v", args[0])
	}
	return "test-key-default"
}

func (f *TestValueFetcher) FetchValue(args ...any) (string, error) {
	// 模拟网络延迟
	time.Sleep(time.Millisecond * 10)
	
	// 增加计数器，用于验证并发调用
	atomic.AddInt64(&f.counter, 1)
	
	// 模拟偶尔的失败
	if atomic.LoadInt64(&f.counter)%10 == 0 {
		return "", errors.New("simulated error")
	}
	
	return fmt.Sprintf("value-%d", atomic.LoadInt64(&f.counter)), nil
}

// TestDefaultValueFetcher 用于测试的 DefaultValueFetcher 实现
type TestDefaultValueFetcher struct {
	TestValueFetcher
}

func (f *TestDefaultValueFetcher) DefaultValue(args ...any) (string, error) {
	return "default-value", nil
}

// TestValueFetcherWithDefaultError 返回 UseDefaultValue 错误的 fetcher
type TestValueFetcherWithDefaultError struct {
	TestValueFetcher
}

func (f *TestValueFetcherWithDefaultError) FetchValue(args ...any) (string, error) {
	atomic.AddInt64(&f.counter, 1)
	return "", UseDefaultValue
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	fetcher := &TestValueFetcher{}
	config := NewCacheCfg[string](time.Second*5, false)
	config.ValueFetcher = fetcher

	var wg sync.WaitGroup
	numGoroutines := 100
	numCalls := 10

	// 启动多个 goroutine 并发访问
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				key := fmt.Sprintf("key-%d", id)
				_, err := config.GetValue(key)
				if err != nil && !errors.Is(err, errUseOutdatedValue) {
					t.Errorf("Unexpected error for goroutine %d, call %d: %v", id, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Completed %d concurrent calls from %d goroutines", numGoroutines*numCalls, numGoroutines)
}

// TestConcurrentAccessWithForceUpdate 测试 ForceUpdate 模式下的并发访问
func TestConcurrentAccessWithForceUpdate(t *testing.T) {
	fetcher := &TestValueFetcher{}
	config := NewCacheCfg[string](time.Second*5, true) // ForceUpdate = true
	config.ValueFetcher = fetcher

	var wg sync.WaitGroup
	numGoroutines := 50
	numCalls := 20

	// 启动多个 goroutine 并发访问
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				key := fmt.Sprintf("force-key-%d", id)
				_, err := config.GetValue(key)
				if err != nil {
					t.Errorf("Unexpected error for goroutine %d, call %d: %v", id, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Completed %d concurrent calls with ForceUpdate from %d goroutines", numGoroutines*numCalls, numGoroutines)
}

// TestConcurrentAccessWithDefaultValue 测试使用默认值的并发访问
func TestConcurrentAccessWithDefaultValue(t *testing.T) {
	fetcher := &TestValueFetcherWithDefaultError{}
	config := NewCacheCfg[string](time.Second*5, false)
	config.ValueFetcher = fetcher

	var wg sync.WaitGroup
	numGoroutines := 30
	numCalls := 15

	// 启动多个 goroutine 并发访问
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				key := fmt.Sprintf("default-key-%d", id)
				value, err := config.GetValue(key)
				if err != nil {
					t.Errorf("Unexpected error for goroutine %d, call %d: %v", id, j, err)
				}
				if value != "default-value" {
					t.Errorf("Expected default value, got: %s", value)
				}
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Completed %d concurrent calls with default values from %d goroutines", numGoroutines*numCalls, numGoroutines)
}

// TestConcurrentAccessWithAutoClean 测试自动清理模式下的并发访问
func TestConcurrentAccessWithAutoClean(t *testing.T) {
	fetcher := &TestValueFetcher{}
	config := NewCacheCfgWithAutoClean[string](time.Millisecond*100, false, time.Millisecond*50)
	config.ValueFetcher = fetcher
	defer config.StopCleaner()

	var wg sync.WaitGroup
	numGoroutines := 20
	numCalls := 30

	// 启动多个 goroutine 并发访问
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				key := fmt.Sprintf("clean-key-%d", id)
				_, err := config.GetValue(key)
				if err != nil && !errors.Is(err, errUseOutdatedValue) {
					t.Errorf("Unexpected error for goroutine %d, call %d: %v", id, j, err)
				}
				// 添加一些延迟，让清理器有机会运行
				time.Sleep(time.Millisecond * 5)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Completed %d concurrent calls with auto-clean from %d goroutines", numGoroutines*numCalls, numGoroutines)
}

// TestRaceCondition 专门测试之前修复的竞态条件
func TestRaceCondition(t *testing.T) {
	// 创建一个会频繁失败的 fetcher，触发使用过期值的逻辑
	fetcher := &TestValueFetcher{}
	config := NewCacheCfg[string](time.Millisecond*10, false) // 很短的 TTL
	config.ValueFetcher = fetcher

	// 先获取一个值，让它过期
	_, _ = config.GetValue("race-test-key")
	time.Sleep(time.Millisecond * 20) // 等待过期

	var wg sync.WaitGroup
	numGoroutines := 100
	numCalls := 5

	// 启动大量 goroutine 并发访问过期的缓存
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				_, err := config.GetValue("race-test-key")
				// 这里应该不会 panic，即使返回错误也是预期的
				if err != nil && !errors.Is(err, errUseOutdatedValue) && !errors.Is(err, errors.New("simulated error")) {
					t.Errorf("Unexpected error for goroutine %d, call %d: %v", id, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Completed race condition test with %d concurrent calls from %d goroutines", numGoroutines*numCalls, numGoroutines)
}

// TestMixedOperations 测试混合操作的并发安全性
func TestMixedOperations(t *testing.T) {
	fetcher := &TestValueFetcher{}
	config := NewCacheCfg[string](time.Second*2, false)
	config.ValueFetcher = fetcher

	var wg sync.WaitGroup
	numGoroutines := 50

	// 启动多个 goroutine，每个执行不同的操作
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// 混合使用不同的 key
			keys := []string{
				fmt.Sprintf("mixed-key-%d", id),
				fmt.Sprintf("mixed-key-%d-alt", id),
				"shared-key", // 共享的 key，更容易触发竞态条件
			}
			
			for j := 0; j < 20; j++ {
				key := keys[j%len(keys)]
				_, err := config.GetValue(key)
				if err != nil && !errors.Is(err, errUseOutdatedValue) {
					t.Errorf("Unexpected error for goroutine %d, call %d: %v", id, j, err)
				}
				
				// 随机延迟
				if j%5 == 0 {
					time.Sleep(time.Millisecond * time.Duration(id%10))
				}
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Completed mixed operations test with %d goroutines", numGoroutines)
}

// BenchmarkConcurrentAccess 并发访问的性能基准测试
func BenchmarkConcurrentAccess(b *testing.B) {
	fetcher := &TestValueFetcher{}
	config := NewCacheCfg[string](time.Second*5, false)
	config.ValueFetcher = fetcher

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-%d", counter%100)
			_, _ = config.GetValue(key)
			counter++
		}
	})
} 