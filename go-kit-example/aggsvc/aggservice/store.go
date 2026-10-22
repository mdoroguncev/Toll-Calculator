package aggservice

import (
	"fmt"

	"github.com/mdoroguncev/toll-calculator/types"
)

type MemoryStore struct {
	data map[int]float64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[int]float64),
	}
}

func (ms *MemoryStore) Insert(dist types.Distance) error {
	ms.data[dist.OBUID] += dist.Value
	return nil
}

func (ms *MemoryStore) Get(id int) (float64, error) {
	dist, ok := ms.data[id]
	if !ok {
		return 0.0, fmt.Errorf("could not find distance for obu id %d", id)
	}
	return dist, nil
}
