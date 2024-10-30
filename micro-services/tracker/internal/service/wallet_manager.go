package service

import (
    "context"
    "encoding/json"
    "sync"
    "time"

    "github.com/say8hi/walletTracker/internal/domain/models"
    "github.com/say8hi/walletTracker/internal/domain/repositories"
    "github.com/say8hi/walletTracker/pkg/logger"
    "github.com/say8hi/walletTracker/pkg/metrics"
)

type WalletManager struct {
    subscriptions  map[string]*models.Subscription
    solanaClient   repositories.SolanaClient
    publisher      repositories.TransactionPublisher
    logger        logger.Logger
    metrics       *metrics.Metrics
    mu            sync.RWMutex
    ctxMap        map[string]context.CancelFunc
}

func NewWalletManager(
    solanaClient repositories.SolanaClient,
    publisher repositories.TransactionPublisher,
    logger logger.Logger,
    metrics *metrics.Metrics,
) *WalletManager {
    return &WalletManager{
        subscriptions: make(map[string]*models.Subscription),
        solanaClient:  solanaClient,
        publisher:     publisher,
        logger:       logger,
        metrics:      metrics,
        ctxMap:       make(map[string]context.CancelFunc),
    }
}

func (wm *WalletManager) AddWallet(address string, chatID int64) error {
    wm.mu.Lock()
    defer wm.mu.Unlock()

    sub, exists := wm.subscriptions[address]
    if !exists {
        sub = &models.Subscription{
            Address: address,
            ChatIDs: make(map[int64]struct{}),
        }
        wm.subscriptions[address] = sub

        if err := wm.startWalletTracking(address); err != nil {
            delete(wm.subscriptions, address)
            return err
        }

        wm.metrics.ActiveSubscriptions.Inc()
    }

    sub.ChatIDs[chatID] = struct{}{}
    wm.metrics.SubscriberCount.WithLabelValues(address).Inc()

    wm.logger.Info("Added wallet tracking",
        "address", address,
        "chat_id", chatID,
        "total_subscribers", len(sub.ChatIDs),
    )

    wm.metrics.ActiveWallets.WithLabelValues(address).Set(1)
    wm.metrics.SubscribersCount.WithLabelValues(address).Set(float64(len(sub.ChatIDs)))

    return nil
}

func (wm *WalletManager) RemoveWallet(address string, chatID int64) error {
    wm.mu.Lock()
    defer wm.mu.Unlock()

    sub, exists := wm.subscriptions[address]
    if !exists {
        return nil
    }

    delete(sub.ChatIDs, chatID)
    wm.metrics.SubscriberCount.WithLabelValues(address).Dec()
    if len(sub.ChatIDs) == 0 {
        wm.metrics.ActiveWallets.WithLabelValues(address).Set(0)
        wm.metrics.SubscribersCount.WithLabelValues(address).Set(0)
    } else {
        wm.metrics.SubscribersCount.WithLabelValues(address).Set(float64(len(sub.ChatIDs)))
    }

    if len(sub.ChatIDs) == 0 {
        wm.stopWalletTracking(address)
        delete(wm.subscriptions, address)
        wm.metrics.ActiveSubscriptions.Dec()

        wm.logger.Info("Removed wallet tracking completely", "address", address)
    } else {
        wm.logger.Info("Removed subscriber from wallet",
            "address", address,
            "chat_id", chatID,
            "remaining_subscribers", len(sub.ChatIDs),
        )
    }

    return nil
}

func (wm *WalletManager) startWalletTracking(address string) error {
    ctx, cancel := context.WithCancel(context.Background())
    wm.ctxMap[address] = cancel

    txChan, err := wm.solanaClient.SubscribeToWallet(ctx, address)
    if err != nil {
        cancel()
        delete(wm.ctxMap, address)
        wm.metrics.SubscriptionErrors.WithLabelValues(address, "subscription").Inc()
        return err
    }

    go wm.handleWalletTransactions(ctx, address, txChan)

    wm.logger.Info("Started tracking wallet", "address", address)
    return nil
}

func (wm *WalletManager) stopWalletTracking(address string) {
    if cancel, exists := wm.ctxMap[address]; exists {
        cancel()
        delete(wm.ctxMap, address)
    }
}

func (wm *WalletManager) handleWalletTransactions(ctx context.Context, address string, txChan <-chan *models.Transaction) {
    for {
        select {
        case <-ctx.Done():
            return
        case tx, ok := <-txChan:
            if !ok {
                wm.logger.Warn("Transaction channel closed", "address", address)
                return
            }

            start := time.Now()

            wm.mu.RLock()
            sub, exists := wm.subscriptions[address]
            if !exists {
                wm.mu.RUnlock()
                return
            }

            chatIDs := make([]int64, 0, len(sub.ChatIDs))
            for chatID := range sub.ChatIDs {
                chatIDs = append(chatIDs, chatID)
            }
            wm.mu.RUnlock()

            notification := struct {
                *models.Transaction
                ChatIDs []int64 `json:"chat_ids"`
            }{
                Transaction: tx,
                ChatIDs:    chatIDs,
            }

            if err := wm.publisher.PublishTransaction(ctx, notification); err != nil {
                wm.logger.Error("Failed to publish transaction",
                    "error", err,
                    "address", address,
                )
                wm.metrics.CommandErrors.WithLabelValues(address, "publish").Inc()
                continue
            }

            wm.metrics.TransactionsProcessed.WithLabelValues(address).Inc()
            wm.metrics.ProcessingLatency.WithLabelValues(address).
                Observe(float64(time.Since(start).Milliseconds()))
            wm.metrics.TransactionsByWallet.WithLabelValues(address).Inc()
            wm.metrics.LastTransactionTime.WithLabelValues(address).Set(float64(time.Now().Unix()))
        }
    }
}

func (wm *WalletManager) ProcessCommand(ctx context.Context, cmdStr string) error {
    var cmd models.WalletCommand
    if err := json.Unmarshal([]byte(cmdStr), &cmd); err != nil {
        wm.metrics.CommandErrors.WithLabelValues("parse", "json").Inc()
        return err
    }

    start := time.Now()

    var err error
    switch cmd.Action {
    case "add":
        err = wm.AddWallet(cmd.Address, cmd.ChatID)
    case "remove":
        err = wm.RemoveWallet(cmd.Address, cmd.ChatID)
    default:
        wm.logger.Error("Unknown command action", "action", cmd.Action)
        wm.metrics.CommandErrors.WithLabelValues("unknown", "action").Inc()
        return nil
    }

    if err != nil {
        wm.metrics.CommandErrors.WithLabelValues(cmd.Action, "processing").Inc()
        return err
    }

    wm.metrics.CommandsProcessed.WithLabelValues(cmd.Action).Inc()
    wm.metrics.ProcessingLatency.WithLabelValues("command").
        Observe(float64(time.Since(start).Milliseconds()))

    return nil
}

func (wm *WalletManager) ListenForCommands(ctx context.Context, cmdChan <-chan string) {
    for {
        select {
        case <-ctx.Done():
            return
        case cmd, ok := <-cmdChan:
            if !ok {
                return
            }
            if err := wm.ProcessCommand(ctx, cmd); err != nil {
                wm.logger.Error("Failed to process command", "error", err)
            }
        }
    }
}

func (wm *WalletManager) Shutdown(ctx context.Context) error {
    wm.mu.Lock()
    defer wm.mu.Unlock()

    for address := range wm.ctxMap {
        wm.stopWalletTracking(address)
    }

    if err := wm.solanaClient.Close(); err != nil {
        return err
    }

    return wm.publisher.Close()
}
