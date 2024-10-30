package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	// Transactions
	TransactionsProcessed *prometheus.CounterVec
	ProcessingLatency     *prometheus.HistogramVec

	// Subscriptions
	ActiveSubscriptions prometheus.Gauge
	SubscriptionErrors  *prometheus.CounterVec
	SubscriberCount     *prometheus.GaugeVec

	// WebSocket
	WebsocketReconnects *prometheus.CounterVec
	WebsocketLatency    *prometheus.HistogramVec

	// Redis
	RedisPublishLatency *prometheus.HistogramVec
	RedisErrors         *prometheus.CounterVec

	// Commands
	CommandsProcessed *prometheus.CounterVec
	CommandErrors     *prometheus.CounterVec

	ActiveWallets        *prometheus.GaugeVec
	SubscribersCount     *prometheus.GaugeVec
	TransactionsByWallet *prometheus.CounterVec
	LastTransactionTime  *prometheus.GaugeVec
}

func NewMetrics(namespace string) *Metrics {
	safeNamespace := strings.ReplaceAll(namespace, "-", "_")

	return &Metrics{
		TransactionsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: safeNamespace,
				Name:      "transactions_processed_total",
				Help:      "The total number of processed transactions",
			},
			[]string{"wallet_address"},
		),

		ProcessingLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: safeNamespace,
				Name:      "transaction_processing_latency_ms",
				Help:      "Transaction processing latency in milliseconds",
				Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"wallet_address"},
		),

		ActiveSubscriptions: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: safeNamespace,
				Name:      "active_subscriptions",
				Help:      "The number of currently active wallet subscriptions",
			},
		),

		SubscriptionErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: safeNamespace,
				Name:      "subscription_errors_total",
				Help:      "The total number of subscription errors",
			},
			[]string{"wallet_address", "error_type"},
		),

		SubscriberCount: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: safeNamespace,
				Name:      "subscribers_count",
				Help:      "The number of subscribers per wallet",
			},
			[]string{"wallet_address"},
		),

		WebsocketReconnects: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: safeNamespace,
				Name:      "websocket_reconnects_total",
				Help:      "The total number of websocket reconnections",
			},
			[]string{"wallet_address"},
		),

		WebsocketLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: safeNamespace,
				Name:      "websocket_latency_ms",
				Help:      "WebSocket message latency in milliseconds",
				Buckets:   []float64{1, 5, 10, 25, 50, 100},
			},
			[]string{"wallet_address"},
		),

		RedisPublishLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: safeNamespace,
				Name:      "redis_publish_latency_ms",
				Help:      "Redis publish latency in milliseconds",
				Buckets:   []float64{1, 5, 10, 25, 50},
			},
			[]string{"channel"},
		),

		RedisErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: safeNamespace,
				Name:      "redis_errors_total",
				Help:      "The total number of Redis errors",
			},
			[]string{"operation", "error_type"},
		),

		CommandsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: safeNamespace,
				Name:      "commands_processed_total",
				Help:      "The total number of processed commands",
			},
			[]string{"command_type"},
		),

		CommandErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: safeNamespace,
				Name:      "command_errors_total",
				Help:      "The total number of command processing errors",
			},
			[]string{"command_type", "error_type"},
		),
		ActiveWallets: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: safeNamespace,
				Name:      "wallet_tracking_status",
				Help:      "Status of wallet tracking (1 = active, 0 = inactive)",
			},
			[]string{"wallet_address"},
		),

		SubscribersCount: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: safeNamespace,
				Name:      "wallet_subscribers_count",
				Help:      "Number of subscribers per wallet",
			},
			[]string{"wallet_address"},
		),

		TransactionsByWallet: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: safeNamespace,
				Name:      "wallet_transactions_total",
				Help:      "Total number of transactions per wallet",
			},
			[]string{"wallet_address"},
		),

		LastTransactionTime: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: safeNamespace,
				Name:      "wallet_last_transaction_timestamp",
				Help:      "Timestamp of the last transaction per wallet",
			},
			[]string{"wallet_address"},
		),
	}
}
