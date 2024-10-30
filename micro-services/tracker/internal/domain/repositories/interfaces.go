package repositories

import (
    "context"
    "github.com/say8hi/walletTracker/internal/domain/models"
)

type TransactionPublisher interface {
    PublishTransaction(ctx context.Context, data interface{}) error
    Close() error
}

type SolanaClient interface {
    SubscribeToWallet(ctx context.Context, address string) (<-chan *models.Transaction, error)
    Close() error
}
