package cachecfg

import (
	"fmt"
	"sync"
	"time"
)

// RaceDetectorTest 专门用于检测数据竞争的程序
func RaceDetectorTest() {
	fmt.Println("Starting race detector test...")
	
	// 创建一个会频繁失败的 fetcher，更容易触发竞态条件
	fetcher := &TestValueFetcher{}
	config := NewCacheCfg[string](time.Millisecond*5, false) // 很短的 TTL
	config.ValueFetcher = fetcher

	var wg sync.WaitGroup
	numGoroutines := 200
	numCalls := 50

	// 启动大量 goroutine 并发访问
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				key := fmt.Sprintf("race-key-%d", id%10) // 使用较少的 key，增加竞争
				_, err := config.GetValue(key)
				if err != nil {
					// 忽略预期的错误
					_ = err
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("Completed race detector test with %d concurrent calls from %d goroutines\n", 
		numGoroutines*numCalls, numGoroutines)
}

// StressTest 压力测试，模拟高并发场景
func StressTest() {
	fmt.Println("Starting stress test...")
	
	fetcher := &TestValueFetcher{}
	config := NewCacheCfgWithAutoClean[string](time.Millisecond*100, false, time.Millisecond*50)
	config.ValueFetcher = fetcher
	defer config.StopCleaner()

	var wg sync.WaitGroup
	numGoroutines := 500
	numCalls := 100

	// 启动大量 goroutine 进行压力测试
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				key := fmt.Sprintf("stress-key-%d", (id+j)%20) // 使用较少的 key
				_, err := config.GetValue(key)
				if err != nil {
					// 忽略预期的错误
					_ = err
				}
				
				// 随机延迟，模拟真实场景
				if j%10 == 0 {
					time.Sleep(time.Microsecond * time.Duration(id%100))
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("Completed stress test with %d concurrent calls from %d goroutines\n", 
		numGoroutines*numCalls, numGoroutines)
}

// MixedStressTest 混合压力测试，包含多种场景
func MixedStressTest() {
	fmt.Println("Starting mixed stress test...")
	
	// 测试不同的配置
	configs := []*CachableConfig[string]{
		NewCacheCfg[string](time.Millisecond*10, false),
		NewCacheCfg[string](time.Millisecond*10, true),
		NewCacheCfgWithAutoClean[string](time.Millisecond*50, false, time.Millisecond*25),
	}
	
	// 确保清理器被停止
	defer func() {
		for _, config := range configs {
			if config.stopChan != nil {
				config.StopCleaner()
			}
		}
	}()

	var wg sync.WaitGroup
	numGoroutines := 100
	numCalls := 200

	// 为每个配置启动 goroutine
	for configIndex, config := range configs {
		fetcher := &TestValueFetcher{}
		config.ValueFetcher = fetcher
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(configIdx, id int) {
				defer wg.Done()
				for j := 0; j < numCalls; j++ {
					key := fmt.Sprintf("mixed-config-%d-key-%d", configIdx, id%5)
					_, err := configs[configIdx].GetValue(key)
					if err != nil {
						// 忽略预期的错误
						_ = err
					}
				}
			}(configIndex, i)
		}
	}

	wg.Wait()
	fmt.Printf("Completed mixed stress test with %d configurations, %d goroutines each\n", 
		len(configs), numGoroutines)
} 