package redis

import (
    "context"
    "encoding/json"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/say8hi/walletTracker/pkg/metrics"
)

type RedisClient struct {
    client  *redis.Client
    pubChan string
    cmdChan string
    metrics *metrics.Metrics
}

func NewRedisClient(url, pubChan, cmdChan string, metrics *metrics.Metrics) (*RedisClient, error) {
    opt, err := redis.ParseURL(url)
    if err != nil {
        return nil, err
    }

    client := redis.NewClient(opt)
    
    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, err
    }

    return &RedisClient{
        client:  client,
        pubChan: pubChan,
        cmdChan: cmdChan,
        metrics: metrics,
    }, nil
}

func (rc *RedisClient) PublishTransaction(ctx context.Context, data interface{}) error {
    start := time.Now()
    
    jsonData, err := json.Marshal(data)
    if err != nil {
        rc.metrics.RedisErrors.WithLabelValues("publish", "marshal").Inc()
        return err
    }

    if err := rc.client.Publish(ctx, rc.pubChan, jsonData).Err(); err != nil {
        rc.metrics.RedisErrors.WithLabelValues("publish", "send").Inc()
        return err
    }

    rc.metrics.RedisPublishLatency.WithLabelValues(rc.pubChan).
        Observe(float64(time.Since(start).Milliseconds()))

    return nil
}

func (rc *RedisClient) SubscribeToCommands(ctx context.Context) (<-chan string, error) {
    pubsub := rc.client.Subscribe(ctx, rc.cmdChan)
    
    if _, err := pubsub.Receive(ctx); err != nil {
        return nil, err
    }

    cmdChan := make(chan string)
    
    go func() {
        defer close(cmdChan)
        defer pubsub.Close()

        for {
            select {
            case <-ctx.Done():
                return
            default:
                msg, err := pubsub.ReceiveMessage(ctx)
                if err != nil {
                    rc.metrics.RedisErrors.WithLabelValues("subscribe", "receive").Inc()
                    continue
                }
                cmdChan <- msg.Payload
            }
        }
    }()

    return cmdChan, nil
}

func (rc *RedisClient) Close() error {
    return rc.client.Close()
}
