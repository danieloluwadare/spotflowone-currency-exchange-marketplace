package repo

import (
	"context"

	"spotflowone/internal/domain"
)

type AllocationRepository interface {
	Add(ctx context.Context, alloc domain.Allocation) error
	ListByPair(ctx context.Context, baseCurrency, quoteCurrency string) ([]domain.Allocation, error)
	ApplyFills(ctx context.Context, baseCurrency, quoteCurrency string, fills []domain.Fill) error
}
