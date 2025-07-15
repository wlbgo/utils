package poslist

import "errors"

type PartialSetList[T any] struct {
	items       []T
	emptySlots  []int
	isEmpty     func(T) bool
	createEmpty func() T
}

// NewPartialSetList - 初始长度为0，支持动态扩展
func NewPartialSetList[T any](isEmpty func(T) bool, createEmpty func() T) (*PartialSetList[T], error) {
	if isEmpty == nil {
		return nil, errors.New("isEmpty is nil")
	}
	if createEmpty == nil {
		return nil, errors.New("createEmpty is nil")
	}
	t := createEmpty()
	if !isEmpty(t) {
		return nil, errors.New("createEmpty() is not empty")
	}
	return &PartialSetList[T]{
		items:       make([]T, 0),
		emptySlots:  make([]int, 0),
		isEmpty:     isEmpty,
		createEmpty: createEmpty,
	}, nil
}

// InsertAt 在指定位置插入元素，如果位置超出当前长度则扩展数组
func (rc *PartialSetList[T]) InsertAt(index int, item T) bool {
	if rc.isEmpty(item) {
		return false
	}
	if index < 0 {
		return false
	}
	// 如果位置超过当前长度，需要扩展数组
	if index >= len(rc.items) {
		newItems := make([]T, index+1)
		copy(newItems, rc.items)
		// 填充新增的空位置到 emptySlots
		for i := len(rc.items); i <= index; i++ {
			newItems[i] = rc.createEmpty() // 零值
			rc.emptySlots = append(rc.emptySlots, i)
		}
		rc.items = newItems
	}
	// 检查指定位置是否为空
	if rc.isEmpty(rc.items[index]) {
		rc.items[index] = item
		rc.removeEmptySlot(index)
		return true
	}
	return false
}

// InsertFirstEmpty 在第一个空位置插入元素，如果没有空位则扩展
func (rc *PartialSetList[T]) InsertFirstEmpty(item T) bool {
	if rc.isEmpty(item) {
		return false
	}
	if len(rc.emptySlots) == 0 {
		// 扩展一个新空位
		rc.items = append(rc.items, rc.createEmpty())
		rc.emptySlots = append(rc.emptySlots, len(rc.items)-1)
	}
	index := rc.emptySlots[0]
	rc.items[index] = item
	rc.removeEmptySlot(index)
	return true
}

// EmptySlotsCount 返回空位置的数量
func (rc *PartialSetList[T]) EmptySlotsCount() int {
	return len(rc.emptySlots)
}

// IsEmpty 检查指定位置是否为空
func (rc *PartialSetList[T]) IsEmpty(index int) bool {
	if index < 0 || index >= len(rc.items) {
		return true
	}
	return rc.isEmpty(rc.items[index])
}

// Length 返回当前列表的长度
func (rc *PartialSetList[T]) Length() int {
	return len(rc.items)
}

// Items 返回当前列表的元素
func (rc *PartialSetList[T]) Items() []T {
	return rc.items
}

func (rc *PartialSetList[T]) removeEmptySlot(index int) {
	for i, slot := range rc.emptySlots {
		if slot == index {
			rc.emptySlots = append(rc.emptySlots[:i], rc.emptySlots[i+1:]...)
			break
		}
	}
}
