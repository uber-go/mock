// Copyright 2010 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gomock_test

import (
	"sort"
	"sync"
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/mock/gomock/internal/mock_gomock"
)

func TestCaptor_NoCaptures(t *testing.T) {
	c := gomock.NewCaptor[int]()

	if got := c.Values(); len(got) != 0 {
		t.Errorf("Values() = %v, want empty", got)
	}
	if v, ok := c.TryValue(); ok || v != 0 {
		t.Errorf("TryValue() = (%v, %v), want (0, false)", v, ok)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Value() did not panic on empty captor")
		}
	}()
	_ = c.Value()
}

func TestCaptor_SingleCapture(t *testing.T) {
	c := gomock.NewCaptor[int]()

	if !c.Matches(42) {
		t.Fatalf("Matches(42) = false, want true")
	}
	if got := c.Value(); got != 42 {
		t.Errorf("Value() = %v, want 42", got)
	}
	v, ok := c.TryValue()
	if !ok || v != 42 {
		t.Errorf("TryValue() = (%v, %v), want (42, true)", v, ok)
	}
	if got := c.Values(); len(got) != 1 || got[0] != 42 {
		t.Errorf("Values() = %v, want [42]", got)
	}
}

func TestCaptor_MultipleCaptures(t *testing.T) {
	c := gomock.NewCaptor[string]()

	for _, s := range []string{"a", "b", "c"} {
		if !c.Matches(s) {
			t.Fatalf("Matches(%q) = false, want true", s)
		}
	}

	if got := c.Value(); got != "c" {
		t.Errorf("Value() = %q, want \"c\"", got)
	}

	got := c.Values()
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("Values() len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("Values()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCaptor_WrongTypeDoesNotCapture(t *testing.T) {
	c := gomock.NewCaptor[int]()

	if c.Matches("not an int") {
		t.Errorf("Matches(string) = true, want false for Captor[int]")
	}
	if c.Matches(1.5) {
		t.Errorf("Matches(float64) = true, want false for Captor[int]")
	}
	if got := c.Values(); len(got) != 0 {
		t.Errorf("Values() = %v, want empty after wrong-type calls", got)
	}

	if !c.Matches(7) {
		t.Errorf("Matches(7) = false, want true")
	}
	if got := c.Values(); len(got) != 1 || got[0] != 7 {
		t.Errorf("Values() = %v, want [7]", got)
	}
}

func TestCaptor_String(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"int", gomock.NewCaptor[int]().String(), "captures values of type int"},
		{"string", gomock.NewCaptor[string]().String(), "captures values of type string"},
		{"struct", gomock.NewCaptor[B]().String(), "captures values of type gomock_test.B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("String() = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestCaptor_ValuesReturnsCopy(t *testing.T) {
	c := gomock.NewCaptor[int]()
	c.Matches(1)
	c.Matches(2)

	got := c.Values()
	got[0] = 99

	if v := c.Value(); v != 2 {
		t.Errorf("Value() = %v, want 2", v)
	}
	if again := c.Values(); again[0] != 1 {
		t.Errorf("Values()[0] = %v after mutating prior copy, want 1", again[0])
	}
}

func TestCaptor_Concurrent(t *testing.T) {
	const n = 200
	c := gomock.NewCaptor[int]()

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(v int) {
			defer wg.Done()
			if !c.Matches(v) {
				t.Errorf("Matches(%d) = false, want true", v)
			}
		}(i)
	}
	wg.Wait()

	got := c.Values()
	if len(got) != n {
		t.Fatalf("Values() len = %d, want %d", len(got), n)
	}

	sort.Ints(got)
	for i := 0; i < n; i++ {
		if got[i] != i {
			t.Errorf("sorted Values()[%d] = %d, want %d", i, got[i], i)
		}
	}
}

func TestCaptor_WithMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := mock_gomock.NewMockMatcher(ctrl)
	c := gomock.NewCaptor[string]()

	mock.EXPECT().Matches(c).Return(true)

	if !mock.Matches("hello") {
		t.Errorf("mock.Matches(\"hello\") = false, want true")
	}
	if got := c.Value(); got != "hello" {
		t.Errorf("Captor.Value() = %q, want \"hello\"", got)
	}
}
