package rank

import (
	"fmt"
	"time"
)

type MockValueFetcher[T any] struct{}

func (m MockValueFetcher[T]) Key(args ...any) string {
	return "mock_key"
}

func (m MockValueFetcher[T]) FetchValue(args ...any) (T, error) {
	return "mock_value", nil
}

func ExampleCachableConfig_GetValue() {
	c := CachableConfig[string]{
		ValueFetcher: &MockValueFetcher[string]{},
		TTL:          5 * time.Minute,
		Cache:        make(map[string]*singleCache[string]),
		Mutex:        sync.RWMutex{},
		ForceUpdate:  false,
	}

	value, err := c.GetValue("arg1", "arg2")
	fmt.Println(value, err)
	// Output: mock_value <nil>
}
