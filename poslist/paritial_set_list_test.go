package poslist

import (
	"gotest.tools/assert"
	"testing"
)

func TestAppendPartialSetList(t *testing.T) {
	type args[T any] struct {
		t          T
		list       []T
		i          *int
		checkEmpty func(*T) bool
	}
	type testCase[T any] struct {
		name    string
		args    args[T]
		wantErr bool
	}
	list := make([]int, 8)
	i := 0
	checkEmpty := func(i *int) bool { return *i == 0 }

	tests := []testCase[int]{
		// TODO: Add test cases.
		{"TestAppendPartialSetList", args[int]{1, list, &i, checkEmpty}, false},
		{"TestAppendPartialSetList", args[int]{2, list, &i, checkEmpty}, false},
		{"TestAppendPartialSetList", args[int]{3, list, &i, checkEmpty}, false},
		{"TestAppendPartialSetList", args[int]{4, list, &i, checkEmpty}, false},
		{"TestAppendPartialSetList", args[int]{5, list, &i, checkEmpty}, false},
		{"TestAppendPartialSetList", args[int]{6, list, &i, checkEmpty}, false},
		{"TestAppendPartialSetList", args[int]{7, list, &i, checkEmpty}, false},
		{"TestAppendPartialSetList", args[int]{8, list, &i, checkEmpty}, false},
		{"TestAppendPartialSetList", args[int]{9, list, &i, checkEmpty}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if list, err := AppendPartialSetList(tt.args.t, tt.args.list, tt.args.i, tt.args.checkEmpty); (err != nil) != tt.wantErr {
				t.Errorf("AppendPartialSetList() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("list: %v", list)
			}
		})
	}
}

func TestSetPartialSetList(t *testing.T) {
	type args[T any] struct {
		t          T
		list       []T
		i          int
		checkEmpty func(*T) bool
	}
	type testCase[T any] struct {
		name    string
		args    args[T]
		wantErr bool
	}

	list := make([]int, 8)
	checkEmpty := func(i *int) bool { return *i == 0 }
	tests := []testCase[int]{
		{"TestSetPartialSetList", args[int]{1, list, 1, checkEmpty}, false},
		{"TestSetPartialSetList", args[int]{4, list, 4, checkEmpty}, false},
		{"TestSetPartialSetList", args[int]{4, list, 4, checkEmpty}, true},
		{"TestSetPartialSetList", args[int]{9, list, 9, checkEmpty}, true},
		{"TestSetPartialSetList", args[int]{-1, list, -1, checkEmpty}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetPartialSetList(tt.args.t, tt.args.list, tt.args.i, tt.args.checkEmpty); (err != nil) != tt.wantErr {
				t.Errorf("SetPartialSetList() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("list: %v", list)
			}
		})
	}
}

func TestTwoFunc(t *testing.T) {
	list := make([]int, 8)
	i := 0
	checkEmpty := func(i *int) bool { return *i == 0 }
	var err error
	// test append
	list, err = AppendPartialSetList(1, list, &i, checkEmpty)
	assert.Assert(t, err == nil)

	list, err = AppendPartialSetList(2, list, &i, checkEmpty)
	assert.Assert(t, err == nil)

	list, err = AppendPartialSetList(3, list, &i, checkEmpty)
	assert.Assert(t, err == nil && i == 3)

	err = SetPartialSetList(4, list, 4, checkEmpty)
	assert.Assert(t, err == nil && i == 3)

	list, err = AppendPartialSetList(5, list, &i, checkEmpty)
	assert.Assert(t, err == nil && len(list) == 8 && i == 4)

	list, err = AppendPartialSetList(6, list, &i, checkEmpty)
	assert.Assert(t, err == nil && len(list) == 8 && i == 6)

	err = SetPartialSetList(7, list, 6, checkEmpty)
	assert.Assert(t, err == nil)

	err = SetPartialSetList(7, list, 6, checkEmpty)
	assert.Assert(t, err != nil)

	list, err = AppendPartialSetList(8, list, &i, checkEmpty)
	assert.Assert(t, err == nil && len(list) == 8 && i == 8)

	list, err = AppendPartialSetList(9, list, &i, checkEmpty)
	assert.Assert(t, err == nil && len(list) == 9 && i == 9)

	t.Logf("list: %v", list)

}
