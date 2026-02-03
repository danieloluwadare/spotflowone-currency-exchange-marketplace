package repo

import (
	"context"
	"errors"
	"sync"

	"spotflowone/internal/domain"
)

var (
	ErrAllocationNotFound = errors.New("allocation not found")
	ErrInsufficientAmount = errors.New("insufficient allocation amount")
)

type MemoryAllocationRepository struct {
	mu          sync.RWMutex
	allocations []domain.Allocation
}

// NewMemoryAllocationRepository creates a new thread-safe in-memory allocation repository.
// Returns a pointer to the initialized repository.
func NewMemoryAllocationRepository() *MemoryAllocationRepository {
	return &MemoryAllocationRepository{
		allocations: make([]domain.Allocation, 0),
	}
}

// Add stores a new allocation in the repository.
// The allocation is appended to the internal slice in a thread-safe manner.
func (r *MemoryAllocationRepository) Add(_ context.Context, alloc domain.Allocation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.allocations = append(r.allocations, alloc)
	return nil
}

// ListByPair retrieves all allocations matching the given base and quote currency pair.
// Returns a slice of matching allocations, or an empty slice if none are found.
func (r *MemoryAllocationRepository) ListByPair(_ context.Context, baseCurrency, quoteCurrency string) ([]domain.Allocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	matching := make([]domain.Allocation, 0)
	for _, alloc := range r.allocations {
		if alloc.BaseCurrency == baseCurrency && alloc.QuoteCurrency == quoteCurrency {
			matching = append(matching, alloc)
		}
	}
	return matching, nil
}

// ApplyFills reduces the available amount for each allocation based on the provided fills.
// For each fill, it finds the matching allocation by vendor ID and currency pair, then subtracts
// the fill amount from the available balance. Returns an error if an allocation is not found
// or if there is insufficient available amount.
func (r *MemoryAllocationRepository) ApplyFills(_ context.Context, baseCurrency, quoteCurrency string, fills []domain.Fill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, fill := range fills {
		found := false
		for i := range r.allocations {
			alloc := &r.allocations[i]
			if alloc.BaseCurrency != baseCurrency || alloc.QuoteCurrency != quoteCurrency {
				continue
			}
			if alloc.VendorID != fill.VendorID {
				continue
			}
			found = true
			if alloc.Available.Cmp(fill.Amount) < 0 {
				return ErrInsufficientAmount
			}
			alloc.Available = alloc.Available.Sub(fill.Amount).Round(8)
			break
		}
		if !found {
			return ErrAllocationNotFound
		}
	}

	return nil
}
