package poslist

import (
	"errors"
)

func AppendPartialSetList[T any](t T, list []T, i *int, checkEmpty func(*T) bool) ([]T, error) {
	startIndex := 0
	if i != nil {
		startIndex = *i
		if startIndex < 0 || startIndex > len(list) { // 可以放过等于
			return list, errors.New("invalid index")
		}
		defer func() {
			*i = startIndex + 1
		}()
	}

	for startIndex < len(list) && !checkEmpty(&list[startIndex]) {
		startIndex += 1
		continue
	}

	if startIndex >= len(list) { // 前边的逻辑决定了，不可能大于
		list = append(list, t)
		startIndex = len(list) - 1
	} else {
		list[startIndex] = t
	}
	return list, nil
}

func SetPartialSetList[T any](t T, list []T, i int, checkEmpty func(*T) bool) error {
	if i < 0 || i >= len(list) {
		return errors.New("invalid index")
	}

	if !checkEmpty(&list[i]) {
		return errors.New("elem is not empty")
	}

	list[i] = t
	return nil
}
