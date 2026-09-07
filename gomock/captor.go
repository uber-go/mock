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

package gomock

import (
	"fmt"
	"reflect"
	"sync"
)

// Captor records argument values passed to a mocked method for later
// inspection. It implements Matcher and matches values of type T;
// values of other types do not match and are not recorded.
//
// Example usage:
//
//	c := gomock.NewCaptor[Order]()
//	mock.EXPECT().Process(c).Return(nil)
//	// run code under test
//	got := c.Value()    // last captured Order
//	all := c.Values()   // all captured Orders
type Captor[T any] struct {
	mu     sync.Mutex
	values []T
}

// NewCaptor returns a Captor that records values of type T.
func NewCaptor[T any]() *Captor[T] {
	return &Captor[T]{}
}

// Matches reports whether x is of type T. On success it records x.
func (c *Captor[T]) Matches(x any) bool {
	v, ok := x.(T)
	if !ok {
		return false
	}
	c.mu.Lock()
	c.values = append(c.values, v)
	c.mu.Unlock()
	return true
}

// String describes what the captor matches.
func (c *Captor[T]) String() string {
	return fmt.Sprintf("captures values of type %s", reflect.TypeOf((*T)(nil)).Elem())
}

// Value returns the most recently captured value. It panics if no value
// has been captured.
func (c *Captor[T]) Value() T {
	v, ok := c.TryValue()
	if !ok {
		panic("gomock: Captor has no captured values")
	}
	return v
}

// TryValue returns the most recently captured value and true, or the zero
// value of T and false if no value has been captured.
func (c *Captor[T]) TryValue() (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.values) == 0 {
		var zero T
		return zero, false
	}
	return c.values[len(c.values)-1], true
}

// Values returns a copy of all captured values in capture order.
func (c *Captor[T]) Values() []T {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]T, len(c.values))
	copy(out, c.values)
	return out
}
