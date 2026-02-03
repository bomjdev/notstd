package notstd

import (
	"errors"
	"testing"
	"time"
)

func TestCacheValue(t *testing.T) {
	t.Run("basic set and get", func(t *testing.T) {
		cv := NewCacheValue[string](0, nil)

		// Initially empty
		if _, ok := cv.GetNoDefault(); ok {
			t.Error("expected empty cache")
		}

		// Set and get
		cv.Set("hello")
		if val, ok := cv.GetNoDefault(); !ok || val != "hello" {
			t.Errorf("expected 'hello', got %v, %v", val, ok)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		cv := NewCacheValue[string](100*time.Millisecond, nil)

		cv.Set("test")
		if val, ok := cv.GetNoDefault(); !ok || val != "test" {
			t.Error("expected value before timeout")
		}

		time.Sleep(150 * time.Millisecond)

		if _, ok := cv.GetNoDefault(); ok {
			t.Error("expected value to be expired")
		}
	})

	t.Run("default function", func(t *testing.T) {
		cv := NewCacheValue[int](0, func() (int, error) {
			return 42, nil
		})

		// GetDefault should call defaultFn
		val, ok, err := cv.GetDefault()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected ok=false for default value")
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}

		// Now it should be cached
		val, ok, err = cv.GetDefault()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected ok=true for cached value")
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("delete returns correct status", func(t *testing.T) {
		cv := NewCacheValue[string](0, nil)

		// Delete empty cache
		if cv.Delete() {
			t.Error("expected false when deleting empty cache")
		}

		// Set and delete
		cv.Set("test")
		if !cv.Delete() {
			t.Error("expected true when deleting actual value")
		}

		// Delete again
		if cv.Delete() {
			t.Error("expected false when deleting already deleted value")
		}
	})

	t.Run("update", func(t *testing.T) {
		// Update without defaultFn should error
		cv := NewCacheValue[int](0, nil)
		if err := cv.Update(); err == nil {
			t.Error("expected error when updating without defaultFn")
		}

		// Update with defaultFn
		counter := 0
		cv = NewCacheValue[int](0, func() (int, error) {
			counter++
			return counter * 10, nil
		})

		// First update
		if err := cv.Update(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val, ok := cv.GetNoDefault(); !ok || val != 10 {
			t.Errorf("expected 10, got %v, %v", val, ok)
		}

		// Second update should refresh the value
		if err := cv.Update(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val, ok := cv.GetNoDefault(); !ok || val != 20 {
			t.Errorf("expected 20, got %v, %v", val, ok)
		}

		// Update with error
		cv = NewCacheValue[int](0, func() (int, error) {
			return 0, errors.New("test error")
		})
		if err := cv.Update(); err == nil {
			t.Error("expected error from defaultFn")
		}
	})
}

func TestCache(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		cache := NewCache[string, int](0, 0, nil, nil)

		// Set and get
		cache.Set("one", 1)
		if val, ok := cache.GetNoDefault("one"); !ok || val != 1 {
			t.Errorf("expected 1, got %v, %v", val, ok)
		}

		// Has
		if !cache.Has("one") {
			t.Error("expected key to exist")
		}
		if cache.Has("two") {
			t.Error("expected key to not exist")
		}

		// Delete
		if !cache.Delete("one") {
			t.Error("expected true when deleting actual value")
		}
		if cache.Has("one") {
			t.Error("expected key to be deleted")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		cache := NewCache[string, string](100*time.Millisecond, 0, nil, nil)

		cache.Set("key", "value")
		if !cache.Has("key") {
			t.Error("expected key before timeout")
		}

		time.Sleep(150 * time.Millisecond)

		if cache.Has("key") {
			t.Error("expected key to be expired")
		}
	})

	t.Run("default function", func(t *testing.T) {
		callCount := 0
		cache := NewCache[string, int](0, 0, nil, func(key string) (int, error) {
			callCount++
			if key == "error" {
				return 0, errors.New("test error")
			}
			return len(key), nil
		})

		// GetDefault with successful defaultFn
		val, ok, err := cache.GetDefault("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected ok=false for default value")
		}
		if val != 5 {
			t.Errorf("expected 5, got %v", val)
		}
		if callCount != 1 {
			t.Errorf("expected defaultFn to be called once, got %d", callCount)
		}

		// Second call should use cached value
		val, ok, err = cache.GetDefault("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected ok=true for cached value")
		}
		if val != 5 {
			t.Errorf("expected 5, got %v", val)
		}
		if callCount != 1 {
			t.Errorf("expected defaultFn to still be called once, got %d", callCount)
		}

		// GetDefault with error
		_, _, err = cache.GetDefault("error")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("keyFn and SetValue", func(t *testing.T) {
		type User struct {
			ID   string
			Name string
		}

		cache := NewCache[string, User](0, 0, func(u User) string {
			return u.ID
		}, nil)

		user := User{ID: "123", Name: "John"}
		cache.SetValue(user)

		if val, ok := cache.GetNoDefault("123"); !ok || val.Name != "John" {
			t.Errorf("expected user John, got %v, %v", val, ok)
		}

		// GetKeyOf
		if key := cache.GetKeyOf(user); key != "123" {
			t.Errorf("expected key '123', got %v", key)
		}
	})

	t.Run("Keys, Values, Range", func(t *testing.T) {
		cache := NewCache[string, int](0, 0, nil, nil)

		cache.Set("one", 1)
		cache.Set("two", 2)
		cache.Set("three", 3)

		// Keys
		keys := cache.Keys()
		if len(keys) != 3 {
			t.Errorf("expected 3 keys, got %d", len(keys))
		}

		// Values
		values := cache.Values()
		if len(values) != 3 {
			t.Errorf("expected 3 values, got %d", len(values))
		}

		// Range
		count := 0
		cache.Range(func(k string, v int) bool {
			count++
			return true
		})
		if count != 3 {
			t.Errorf("expected to range over 3 items, got %d", count)
		}

		// Range with early stop
		count = 0
		cache.Range(func(k string, v int) bool {
			count++
			return count < 2
		})
		if count != 2 {
			t.Errorf("expected to stop at 2, got %d", count)
		}
	})

	t.Run("SetMany and DeleteMany", func(t *testing.T) {
		cache := NewCache[string, int](0, 0, nil, nil)

		cache.Set("one", 1)

		// SetMany
		overwrites := cache.SetMany(map[string]int{
			"one":   10, // overwrite
			"two":   2,  // new
			"three": 3,  // new
		})
		if overwrites != 1 {
			t.Errorf("expected 1 overwrite, got %d", overwrites)
		}
		if cache.Len() != 3 {
			t.Errorf("expected 3 items, got %d", cache.Len())
		}

		// DeleteMany
		deleted := cache.DeleteMany([]string{"one", "two", "nonexistent"})
		if deleted != 2 {
			t.Errorf("expected 2 deletions, got %d", deleted)
		}
		if cache.Len() != 1 {
			t.Errorf("expected 1 item, got %d", cache.Len())
		}
	})

	t.Run("update", func(t *testing.T) {
		// Update without defaultFn should error
		cache := NewCache[string, int](0, 0, nil, nil)
		if err := cache.Update("key"); err == nil {
			t.Error("expected error when updating without defaultFn")
		}

		// Update with defaultFn
		cache = NewCache[string, int](0, 0, nil, func(key string) (int, error) {
			return len(key) * 10, nil
		})

		// Update should create new entry
		if err := cache.Update("hello"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val, ok := cache.GetNoDefault("hello"); !ok || val != 50 {
			t.Errorf("expected 50, got %v, %v", val, ok)
		}

		// Update existing entry should refresh
		cache.Set("hello", 999)
		if err := cache.Update("hello"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val, ok := cache.GetNoDefault("hello"); !ok || val != 50 {
			t.Errorf("expected 50, got %v, %v", val, ok)
		}

		// Update with error
		cache = NewCache[string, int](0, 0, nil, func(key string) (int, error) {
			if key == "error" {
				return 0, errors.New("test error")
			}
			return 42, nil
		})
		if err := cache.Update("error"); err == nil {
			t.Error("expected error from defaultFn")
		}
	})
}
