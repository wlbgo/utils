`diststat` 基于 Redis 分布式统计，用于在分布式环境中高效地统计计数器数据。

相比于基于Prometheus指标统计，有不依赖于指标平台、数据量小、实时性强、可灵活获取数据的优势。

它提供了一个简单的接口来增加计数器，并定期将计数器数据刷新到 Redis 中。

使用并修改Python脚本，可以针对不同维度统计出CSV表格数据。再结合 [企业微信机器人](https://github.com/easy-wx/wecom-bot-svr?tab=readme-ov-file#5-%E5%8F%91%E9%80%81%E6%96%87%E4%BB%B6)
等社交软件，可以方便灵活的提供给分析人员。

![fetch_in_wecom.png](images%2Ffetch_in_wecom.png)

![stat_detail.png](images%2Fstat_detail.png)

## 使用方法

使用 `go get` 命令安装 `diststat`：

```sh
go get github.com/yourusername/diststat
```

### 初始化

首先，创建一个 `StatHelper` 实例并进行初始化：

```go
package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/yourusername/diststat"
	"time"
)

func main() {
	rds := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	statHelper := &diststat.StatHelper{
		Rds:           rds,
		StatKeyPrefix: "myapp:stats",
		Period:        1 * time.Minute,
		PeriodStart:   time.Now().Truncate(time.Hour),
		StateKeyTTL:   24 * time.Hour,
		FlushPeriod:   1 * time.Second,
	}

	err := statHelper.Init()
	if err != nil {
		panic(err)
	}

	// 使用计数器
	statHelper.CounterIncr("example_counter")
}
```

### 配置参数

- `Rds`: Redis 客户端实例。
- `StatKeyPrefix`: 统计数据的键前缀。
- `Period`: 统计周期，例如 1 分钟。
- `PeriodStart`: 统计周期的起始时间。
- `StateKeyTTL`: 统计数据在 Redis 中的存活时间。
- `FlushPeriod`: 刷新计数器数据到 Redis 的周期，默认为 1 秒。

### 增加计数器

使用 `CounterIncr` 方法增加计数器：

```go
statHelper.CounterIncr("example_counter")
```

### 统计分布

为了统计 diststat 库产生的数据，使用者可以基于 pylib/stat_dist.py 脚本进行改造。

该脚本从 Redis 中获取统计数据，并将其格式化为 CSV 文件。

当前脚本获取的是一个统计周期内，按照实验、轮次、道具ID的进行统计。

用户可以自己设计 key 前缀和子标签的格式，丰富统计维度。

## 最佳实践

有一个推荐活动，活动会产出共20个左右的道具排序。

有 dev 和 online 两个环境。每个环境有三个实验，每个实验会使用不同算法、根据不同的用户信息产生不同的道具排序。

而道具落到具体的位置，又会有不同的原因，比如新品置顶、固定位置、算法打分、特殊置底等。

预期能统计每个时间段，按照环境、实验维度，统计不同位置的道具、原因的分布情况。

我们可以进行如下的*key前缀和子标签*的设计，达到统计的效果。

- key前缀：`{recommend_app_id}:{online}`
- 子标签：`{experiment_id}:{reason}:{index}:{item_id}`

以上的设计，就和我们提供的 ``[exp_dist_hour_stat_test.go](exp_dist_hour_stat_test.go)`` 文件
以及 ``[stat_dist.py](stat_dist.py)`` 文件相对应。

如果道具数量较多，我们可以将不同实验，放到 key 前缀中，减少子标签的数据量，从而减少每个hash中的元素个数。
但在统计时，需要知道具体有哪些实验ID，以汇总统计。
