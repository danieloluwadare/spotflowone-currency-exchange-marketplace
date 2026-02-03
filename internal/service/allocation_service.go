package service

import (
	"context"
	"strings"

	"github.com/shopspring/decimal"

	"spotflowone/internal/domain"
	"spotflowone/internal/repo"
)

type AllocationService struct {
	repo repo.AllocationRepository
}

// NewAllocationService creates a new allocation service with the provided repository.
// Returns a pointer to the initialized service.
func NewAllocationService(repo repo.AllocationRepository) *AllocationService {
	return &AllocationService{repo: repo}
}

// AddAllocation validates and stores a new vendor allocation.
// It normalizes currency codes, validates inputs (non-empty vendor ID, valid currencies,
// different base/quote currencies, positive rate and available amount), then persists
// the allocation via the repository. Returns ErrInvalidInput if validation fails.
func (s *AllocationService) AddAllocation(ctx context.Context, alloc domain.Allocation) error {
	alloc.BaseCurrency = domain.NormalizeCurrency(alloc.BaseCurrency)
	alloc.QuoteCurrency = domain.NormalizeCurrency(alloc.QuoteCurrency)
	alloc.VendorID = strings.TrimSpace(alloc.VendorID)

	if alloc.VendorID == "" {
		return ErrInvalidInput
	}
	if !domain.IsValidCurrency(alloc.BaseCurrency) || !domain.IsValidCurrency(alloc.QuoteCurrency) {
		return ErrInvalidInput
	}
	if alloc.BaseCurrency == alloc.QuoteCurrency {
		return ErrInvalidInput
	}
	if !isPositive(alloc.Rate) || !isPositive(alloc.Available) {
		return ErrInvalidInput
	}

	return s.repo.Add(ctx, alloc)
}

// isPositive checks if a decimal value is greater than zero.
// Returns true if the value is positive, false otherwise.
func isPositive(value decimal.Decimal) bool {
	return value.Cmp(decimal.Zero) > 0
}
