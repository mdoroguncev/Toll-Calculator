package aggregator

import (
	"context"
	"fmt"
	"sync"

	"github.com/mdoroguncev/toll-calculator/internal/types"
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

func (ms *MemoryStore) Insert(ctx context.Context, dist types.Distance) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.data[dist.OBUID] += dist.Value
	return nil
}

func (ms *MemoryStore) Get(ctx context.Context, id int) (float64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	dist, ok := ms.data[id]
	if !ok {
		return 0.0, fmt.Errorf("obu %d: %w", id, ErrNotFound)
	}
	return dist, nil
}

func NewStore(storeType string) (Store, error) {
	switch storeType {
	case "memory":
		return NewMemoryStore(), nil
	default:
		return nil, fmt.Errorf("unknown store type %q", storeType)
	}
}
