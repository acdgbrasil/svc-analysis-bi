// Package ingestion provides the NATS JetStream consumer adapter.
package ingestion

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSConsumerConfig holds the connection and subscription parameters for
// the NATS JetStream consumer.
type NATSConsumerConfig struct {
	// URL is the NATS server URL (e.g. "nats://localhost:4222").
	URL string

	// StreamName is the JetStream stream to bind to (e.g. "SOCIAL_CARE_EVENTS").
	StreamName string

	// ConsumerName is the durable consumer name for pull-based subscriptions.
	ConsumerName string
}

// natsConsumer implements the Consumer interface using NATS JetStream
// pull-based subscriptions for backpressure control.
type natsConsumer struct {
	cfg NATSConsumerConfig
}

// NewNATSConsumer creates a Consumer that subscribes to a NATS JetStream
// stream using pull-based delivery with a durable consumer name.
func NewNATSConsumer(cfg NATSConsumerConfig) Consumer {
	return &natsConsumer{cfg: cfg}
}

// Subscribe connects to NATS, creates a pull subscription on the configured
// stream, and delivers messages to the provided channel. It blocks until ctx
// is cancelled or a fatal connection error occurs. The caller owns the
// channel and must close it after Subscribe returns.
func (c *natsConsumer) Subscribe(ctx context.Context, out chan<- RawMessage) error {
	nc, err := nats.Connect(c.cfg.URL,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.Name("svc-analysis-bi"),
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConsumerConnectionFailed, err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("%w: jetstream: %v", ErrConsumerConnectionFailed, err)
	}

	// Subscribe with durable consumer. The "social-care.>" filter matches
	// all social-care domain events in the stream.
	sub, err := js.PullSubscribe(
		"social-care.>",
		c.cfg.ConsumerName,
		nats.BindStream(c.cfg.StreamName),
	)
	if err != nil {
		return fmt.Errorf("%w: subscribe: %v", ErrConsumerConnectionFailed, err)
	}
	defer func() {
		_ = sub.Unsubscribe()
	}()

	// Pull loop with context cancellation and bounded fetch.
	const batchSize = 10
	const fetchTimeout = 2 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msgs, err := sub.Fetch(batchSize, nats.MaxWait(fetchTimeout))
		if err != nil {
			if err == nats.ErrTimeout {
				continue // no messages available, poll again
			}
			// Check if context was cancelled during fetch
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("%w: fetch: %v", ErrConsumerConnectionFailed, err)
		}

		for _, msg := range msgs {
			natsMsg := msg // capture for closure
			raw := RawMessage{
				Subject: natsMsg.Subject,
				Data:    natsMsg.Data,
				Ack: func() error {
					return natsMsg.Ack()
				},
			}

			select {
			case out <- raw:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// NATSHealthChecker wraps a NATS connection for readiness probe checks.
// It implements the api.NATSChecker interface.
type NATSHealthChecker struct {
	nc *nats.Conn
}

// NewNATSHealthChecker creates a health checker from an existing NATS
// connection. The connection is managed externally; the checker only
// inspects its status.
func NewNATSHealthChecker(nc *nats.Conn) *NATSHealthChecker {
	return &NATSHealthChecker{nc: nc}
}

// IsConnected returns true if the NATS connection is active.
func (c *NATSHealthChecker) IsConnected() bool {
	return c.nc != nil && c.nc.IsConnected()
}
