package poslist

import (
	"testing"
)

// 测试用的空值检查函数
func isEmptyString(s string) bool {
	return s == ""
}

func isEmptyInt(i int) bool {
	return i == 0
}

// 测试用的创建空值函数
func createEmptyString() string {
	return ""
}

func createEmptyInt() int {
	return 0
}

func TestNewPartialSetList(t *testing.T) {
	list, err := NewPartialSetList[string](isEmptyString, createEmptyString)
	if err != nil {
		t.Errorf("Failed to create PartialSetList: %v", err)
	}

	if list.Length() != 0 {
		t.Errorf("Expected length 0, got %d", list.Length())
	}

	if list.EmptySlotsCount() != 0 {
		t.Errorf("Expected empty slots count 0, got %d", list.EmptySlotsCount())
	}
}

func TestNewPartialSetListWithInvalidParams(t *testing.T) {
	// 测试 nil isEmpty
	_, err := NewPartialSetList[string](nil, createEmptyString)
	if err == nil {
		t.Error("Expected error for nil isEmpty")
	}

	// 测试 nil createEmpty
	_, err = NewPartialSetList[string](isEmptyString, nil)
	if err == nil {
		t.Error("Expected error for nil createEmpty")
	}

	// 测试 createEmpty 返回非空值
	_, err = NewPartialSetList[string](isEmptyString, func() string { return "not empty" })
	if err == nil {
		t.Error("Expected error for createEmpty returning non-empty value")
	}
}

func TestInsertAt(t *testing.T) {
	list, err := NewPartialSetList[string](isEmptyString, createEmptyString)
	if err != nil {
		t.Fatalf("Failed to create PartialSetList: %v", err)
	}

	// 测试在位置0插入
	if !list.InsertAt(0, "first") {
		t.Error("Failed to insert at position 0")
	}

	if list.Length() != 1 {
		t.Errorf("Expected length 1, got %d", list.Length())
	}

	// 测试在位置5插入（超出当前长度）
	if !list.InsertAt(5, "fifth") {
		t.Error("Failed to insert at position 5")
	}

	if list.Length() != 6 {
		t.Errorf("Expected length 6, got %d", list.Length())
	}

	// 验证空位置数量
	if list.EmptySlotsCount() != 4 { // 位置1,2,3,4是空的
		t.Errorf("Expected empty slots count 4, got %d", list.EmptySlotsCount())
	}

	// 测试在已占用位置插入
	if list.InsertAt(0, "another") {
		t.Error("Should not be able to insert at occupied position")
	}
}

func TestInsertFirstEmpty(t *testing.T) {
	list, err := NewPartialSetList[int](isEmptyInt, createEmptyInt)
	if err != nil {
		t.Fatalf("Failed to create PartialSetList: %v", err)
	}

	// 测试在空列表中插入
	if !list.InsertFirstEmpty(10) {
		t.Error("Failed to insert first element")
	}

	if list.Length() != 1 {
		t.Errorf("Expected length 1, got %d", list.Length())
	}

	// 测试在指定位置插入后，再使用 InsertFirstEmpty
	list.InsertAt(5, 50)

	if !list.InsertFirstEmpty(20) {
		t.Error("Failed to insert in first empty slot")
	}

	// 验证插入到了位置0（第一个空位置）
	if list.IsEmpty(0) {
		t.Error("Position 0 should not be empty after insertion")
	}
}

func TestIsEmpty(t *testing.T) {
	list, err := NewPartialSetList[string](isEmptyString, createEmptyString)
	if err != nil {
		t.Fatalf("Failed to create PartialSetList: %v", err)
	}

	// 测试超出范围的位置
	if !list.IsEmpty(5) {
		t.Error("Out of range position should be considered empty")
	}

	// 插入元素后测试
	list.InsertAt(3, "test")

	if list.IsEmpty(3) {
		t.Error("Position 3 should not be empty after insertion")
	}

	if !list.IsEmpty(2) {
		t.Error("Position 2 should be empty")
	}
}

func TestLengthAndEmptySlotsCount(t *testing.T) {
	list, err := NewPartialSetList[int](isEmptyInt, createEmptyInt)
	if err != nil {
		t.Fatalf("Failed to create PartialSetList: %v", err)
	}

	// 初始状态
	if list.Length() != 0 {
		t.Errorf("Expected initial length 0, got %d", list.Length())
	}

	if list.EmptySlotsCount() != 0 {
		t.Errorf("Expected initial empty slots count 0, got %d", list.EmptySlotsCount())
	}

	// 插入元素到位置5（会扩展到长度6）
	list.InsertAt(5, 50)

	if list.Length() != 6 {
		t.Errorf("Expected length 6 after inserting at position 5, got %d", list.Length())
	}

	if list.EmptySlotsCount() != 5 { // 位置0,1,2,3,4是空的
		t.Errorf("Expected empty slots count 5, got %d", list.EmptySlotsCount())
	}
}

func TestEmptySlotsManagement(t *testing.T) {
	list, err := NewPartialSetList[int](isEmptyInt, createEmptyInt)
	if err != nil {
		t.Fatalf("Failed to create PartialSetList: %v", err)
	}

	// 插入一些元素
	list.InsertAt(0, 10)
	list.InsertAt(2, 30)
	list.InsertAt(4, 50)

	// 验证空位置数量
	expectedEmpty := 2 // 位置1,3是空的
	if list.EmptySlotsCount() != expectedEmpty {
		t.Errorf("Expected empty slots count %d, got %d", expectedEmpty, list.EmptySlotsCount())
	}

	// 验证非空位置
	if list.IsEmpty(0) || list.IsEmpty(2) || list.IsEmpty(4) {
		t.Error("Positions 0, 2, 4 should not be empty")
	}

	// 验证空位置
	if !list.IsEmpty(1) || !list.IsEmpty(3) || !list.IsEmpty(5) {
		t.Error("Positions 1, 3, 5 should be empty")
	}
}
