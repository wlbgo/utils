package main

import (
	"fmt"
	"time"
	"github.com/wlbgo/utils/cachecfg"
)

func main() {
	fmt.Println("=== CachableConfig Race Condition Test ===")
	fmt.Println("This program tests for race conditions in CachableConfig")
	fmt.Println("Run with: go run -race race_test_main.go")
	fmt.Println()

	// 运行数据竞争检测测试
	fmt.Println("1. Running race detector test...")
	cachecfg.RaceDetectorTest()
	fmt.Println()

	// 运行压力测试
	fmt.Println("2. Running stress test...")
	cachecfg.StressTest()
	fmt.Println()

	// 运行混合压力测试
	fmt.Println("3. Running mixed stress test...")
	cachecfg.MixedStressTest()
	fmt.Println()

	fmt.Println("=== All tests completed ===")
	fmt.Println("If no race conditions were detected, the fix is working correctly!")
	fmt.Println("Check the output above for any race condition warnings.")
} 