# cachecfg

`cachecfg` 是一个用于缓存配置值的 Go 包。它提供了一个通用的缓存机制，可以根据需要从源头获取值，并在获取失败时使用默认值。

## 使用方法

### 定义接口

首先，定义一个实现 `ValueFetcher` 接口的类型，用于从源头获取值。
可选地，定义一个实现 `DefaultValueFetcher` 接口的类型，用于在获取值失败时提供默认值。

需要注意，并非所有的 `FetchValue` 返回的错误都会使用默认值，只有当返回的错误是 `DefaultValueError` 类型时才会使用默认值。

```go
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/wlbgo/utils/cachecfg"
)

// MyValueFetcher 实现了 ValueFetcher 接口
type MyValueFetcher struct{}

func (f *MyValueFetcher) Key(args ...any) string {
	return fmt.Sprintf("key-%v", args)
}

func (f *MyValueFetcher) FetchValue(args ...any) (string, error) {
	// 模拟获取值的过程
	return "", &cachecfg.DefaultValueError{Err: errors.New("fetch failed")}
}

// MyDefaultValueFetcher 实现了 DefaultValueFetcher 接口
type MyDefaultValueFetcher struct {
	MyValueFetcher
}

func (f *MyDefaultValueFetcher) DefaultValue(args ...any) (string, error) {
	// 提供默认值
	return "default-value", nil
}
```

### 创建缓存配置

创建一个 `CachableConfig` 实例，并设置缓存的 TTL（Time-To-Live）和是否强制更新。

```go
package main

import (
	"fmt"
	"github.com/wlbgo/utils/cachecfg"
	"time"
)

func main() {
	ttl := 5 * time.Minute
	forceUpdate := true

	valueFetcher := &MyDefaultValueFetcher{}
	cacheCfg := cachecfg.NewCacheCfg[string](ttl, forceUpdate)
	cacheCfg.ValueFetcher = valueFetcher

	value, err := cacheCfg.GetValue("example")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Value:", value)
	}
}
```

### Redis 实现

RedisKeyValueFetcher 是一个实现了 ValueFetcher 接口的类型，用于从 Redis 获取值，返回 key-value 的 ``[]byte`` 结果。

### 其他使用方式

可以嵌套使用。场景与实例，待补充


## TODO 

- [ ] 增加自动清理




