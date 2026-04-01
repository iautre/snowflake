package snowflake

import (
	"sync"
	"testing"
)

func TestNextIDConcurrentUnique(t *testing.T) {
	const n = 10000
	sf := NewSnowflake()
	var wg sync.WaitGroup
	list := make([]int64, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id, err := sf.NextID()
			if err != nil {
				t.Errorf("NextID: %v", err)
				return
			}
			list[idx] = id
		}(i)
	}
	wg.Wait()
	seen := make(map[int64]struct{}, n)
	for _, id := range list {
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id: %d", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != n {
		t.Fatalf("want %d unique ids, got %d", n, len(seen))
	}
}

func TestInvalidWorkerConfig(t *testing.T) {
	s := NewSnowflake().WithWorkerId(32)
	_, err := s.NextID()
	if err == nil {
		t.Fatal("expected error for worker id 32")
	}
}

func TestPackageString(t *testing.T) {
	s := String()
	if s == "" {
		t.Fatal("empty string")
	}
}

func TestCustomSnowflake(t *testing.T) {
	s := NewSnowflake().WithTwepoch(1655647200000).WithWorkerId(1).WithDatacenterId(1)
	if _, err := s.NextID(); err != nil {
		t.Fatal(err)
	}
	idStr, err := s.NextString()
	if err != nil {
		t.Fatal(err)
	}
	if idStr == "" {
		t.Fatal("empty NextString")
	}
}
