package solana

import (
    "context"
    "time"

    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/gagliardetto/solana-go/rpc/ws"
    "github.com/say8hi/walletTracker/internal/domain/models"
    "github.com/say8hi/walletTracker/pkg/metrics"
)

type SolanaWSClient struct {
    client     *ws.Client
    commitment rpc.CommitmentType
    metrics    *metrics.Metrics
}

func NewSolanaWSClient(wsURL string, commitment string, m *metrics.Metrics) (*SolanaWSClient, error) {
    client, err := ws.Connect(context.Background(), wsURL)
    if err != nil {
        return nil, err
    }

    var commitmentType rpc.CommitmentType
    switch commitment {
    case "processed":
        commitmentType = rpc.CommitmentProcessed
    case "confirmed":
        commitmentType = rpc.CommitmentConfirmed
    case "finalized":
        commitmentType = rpc.CommitmentFinalized
    default:
        commitmentType = rpc.CommitmentProcessed
    }

    return &SolanaWSClient{
        client:     client,
        commitment: commitmentType,
        metrics:    m,
    }, nil
}

func (sc *SolanaWSClient) SubscribeToWallet(ctx context.Context, address string) (<-chan *models.Transaction, error) {
    pubKey, err := solana.PublicKeyFromBase58(address)
    if err != nil {
        return nil, err
    }

    sub, err := sc.client.AccountSubscribe(
        pubKey,
        sc.commitment,
    )
    if err != nil {
        return nil, err
    }

    txChan := make(chan *models.Transaction)

    go func() {
        defer close(txChan)
        defer sub.Unsubscribe()

        for {
            select {
            case <-ctx.Done():
                return
            default:
                start := time.Now()
                
                notification, err := sub.Recv()
                if err != nil {
                    sc.metrics.WebsocketReconnects.WithLabelValues(address).Inc()
                    continue
                }

                sc.metrics.WebsocketLatency.WithLabelValues(address).
                    Observe(float64(time.Since(start).Milliseconds()))

                tx := &models.Transaction{
                    FromAddr:   address,
                    Timestamp:  time.Now().UTC(),
                    SlotNumber: notification.Context.Slot,
                }

                select {
                case txChan <- tx:
                    sc.metrics.TransactionsProcessed.WithLabelValues(address).Inc()
                case <-ctx.Done():
                    return
                }
            }
        }
    }()

    return txChan, nil
}

func (sc *SolanaWSClient) Close() error {
    if sc.client != nil {
        sc.client.Close()
    }
    return nil
}
