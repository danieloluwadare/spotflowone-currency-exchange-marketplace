package rates

import (
	"context"

	"spotflowone/internal/domain"
)

type Provider interface {
	GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (domain.MarketRate, error)
}
