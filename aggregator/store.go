package main

import (
	"fmt"
	"sync"

	"github.com/mdoroguncev/toll-calculator/types"
)

type MemoryStore struct {
	data map[int]float64
	mu   sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[int]float64),
	}
}

func (ms *MemoryStore) Insert(dist types.Distance) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.data[dist.OBUID] += dist.Value
	return nil
}

func (ms *MemoryStore) Get(id int) (float64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	dist, ok := ms.data[id]
	if !ok {
		return 0.0, fmt.Errorf("could not find distance for obu id %d", id)
	}
	return dist, nil
}
