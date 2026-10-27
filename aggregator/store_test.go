package main

import (
	"sync"
	"testing"

	"github.com/mdoroguncev/toll-calculator/types"
)

func TestMemoryStoreConcurrent(t *testing.T) {
	s := NewMemoryStore()

	wg := sync.WaitGroup{}

	for i := 0; i < 50; i++ {

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Insert(types.Distance{OBUID: 1, Value: 1}); err != nil {
				t.Errorf("error: %v", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			_, _ = s.Get(1)
		}()
	}
	wg.Wait()
	got, err := s.Get(1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 50 {
		t.Errorf("want 50, got %v", got)
	}
}
