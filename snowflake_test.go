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
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := sf.NextID()
			if err != nil {
				t.Errorf("NextID: %v", err)
				return
			}
			list[i] = id
		}()
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

func TestPackageNextString(t *testing.T) {
	s, err := NextString()
	if err != nil {
		t.Fatal(err)
	}
	if s == "" {
		t.Fatal("empty NextString")
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
