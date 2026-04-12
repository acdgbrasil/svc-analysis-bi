package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewNATSConsumer_ReturnsNonNil(t *testing.T) {
	consumer := NewNATSConsumer(nil, NATSConsumerConfig{
		URL:          "nats://localhost:4222",
		StreamName:   "TEST_STREAM",
		ConsumerName: "test-consumer",
	})
	if consumer == nil {
		t.Fatal("NewNATSConsumer returned nil")
	}
}

func TestNATSConsumer_Subscribe_NilConnection(t *testing.T) {
	consumer := NewNATSConsumer(nil, NATSConsumerConfig{
		StreamName:   "TEST_STREAM",
		ConsumerName: "test-consumer",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out := make(chan RawMessage, 1)
	err := consumer.Subscribe(ctx, out)

	if err == nil {
		t.Fatal("expected error for nil connection, got nil")
	}

	if !errors.Is(err, ErrConsumerConnectionFailed) {
		t.Errorf("expected ErrConsumerConnectionFailed, got: %v", err)
	}
}

func TestNATSHealthChecker_NilConnection(t *testing.T) {
	checker := NewNATSHealthChecker(nil)
	if checker.IsConnected() {
		t.Error("expected IsConnected() to return false for nil connection")
	}
}
