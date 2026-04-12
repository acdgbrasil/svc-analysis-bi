---
name: application-expert
description: >
  Expert skill for designing the Ingestion Pipeline in Go.
  The ingestion layer knows WHAT to do: consume events, anonymize, materialize.
  Uses goroutines and channels for pipeline stages. No business logic -- calls domain functions.
  Use when the user mentions: ingestion, pipeline, consumer, event handler, anonymizer, materializer.
user_invocable: true
---

# Application Expert -- Go Ingestion Pipeline

You are the Ingestion Pipeline specialist. This layer orchestrates the flow from raw NATS events to materialized analytical data without containing business logic.

## Core Pattern

### Event Consumer (NATS JetStream)
```go
package ingestion

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

type ConsumerConfig struct {
    NATSUrl      string
    StreamName   string
    DurableName  string
    SubjectFilter string
}

type Consumer struct {
    cfg     ConsumerConfig
    handler EventHandler
    conn    *nats.Conn
    sub     jetstream.ConsumeContext
    logger  *slog.Logger
}

func NewConsumer(cfg ConsumerConfig, handler EventHandler, logger *slog.Logger) *Consumer {
    return &Consumer{cfg: cfg, handler: handler, logger: logger}
}

func (c *Consumer) Start(ctx context.Context) error {
    conn, err := nats.Connect(c.cfg.NATSUrl)
    if err != nil {
        return fmt.Errorf("nats connect: %w", err)
    }
    c.conn = conn

    js, err := jetstream.New(conn)
    if err != nil {
        return fmt.Errorf("jetstream: %w", err)
    }

    cons, err := js.CreateOrUpdateConsumer(ctx, c.cfg.StreamName, jetstream.ConsumerConfig{
        Durable:       c.cfg.DurableName,
        FilterSubject: c.cfg.SubjectFilter,
        AckPolicy:     jetstream.AckExplicitPolicy,
    })
    if err != nil {
        return fmt.Errorf("consumer create: %w", err)
    }

    c.sub, err = cons.Consume(func(msg jetstream.Msg) {
        if err := c.handler.HandleEvent(ctx, msg.Data()); err != nil {
            c.logger.Error("event processing failed", "error", err, "subject", msg.Subject())
            msg.Nak()
            return
        }
        msg.Ack()
    })
    if err != nil {
        return fmt.Errorf("consume: %w", err)
    }

    return nil
}

func (c *Consumer) Stop() {
    if c.sub != nil {
        c.sub.Stop()
    }
    if c.conn != nil {
        c.conn.Drain()
    }
}
```

### Event Handler (Pipeline Orchestrator)
```go
type EventHandler interface {
    HandleEvent(ctx context.Context, data []byte) error
}

type PipelineHandler struct {
    anonymizer   Anonymizer
    materializer Materializer
    eventLog     EventLog
    logger       *slog.Logger
}

func NewPipelineHandler(
    anonymizer Anonymizer,
    materializer Materializer,
    eventLog EventLog,
    logger *slog.Logger,
) *PipelineHandler {
    return &PipelineHandler{
        anonymizer:   anonymizer,
        materializer: materializer,
        eventLog:     eventLog,
        logger:       logger,
    }
}

func (h *PipelineHandler) HandleEvent(ctx context.Context, data []byte) error {
    // 1. Deserialize raw event
    raw, err := deserializeEvent(data)
    if err != nil {
        return fmt.Errorf("deserialize: %w", err)
    }

    // 2. Check idempotency (skip if already processed)
    processed, err := h.eventLog.IsProcessed(ctx, raw.EventID)
    if err != nil {
        return fmt.Errorf("idempotency check: %w", err)
    }
    if processed {
        return nil // Already handled
    }

    // 3. Anonymize (suppress PII, generalize quasi-identifiers)
    anonymized, err := h.anonymizer.Anonymize(ctx, raw)
    if err != nil {
        return fmt.Errorf("anonymize: %w", err)
    }

    // 4. Materialize (upsert into analytical store)
    if err := h.materializer.Materialize(ctx, anonymized); err != nil {
        return fmt.Errorf("materialize: %w", err)
    }

    // 5. Mark as processed (AFTER successful materialization)
    if err := h.eventLog.MarkProcessed(ctx, raw.EventID, raw.EventType); err != nil {
        return fmt.Errorf("mark processed: %w", err)
    }

    return nil
}
```

### Ports (Interfaces)
```go
type Anonymizer interface {
    Anonymize(ctx context.Context, raw domain.RawEvent) (domain.GeneralizedEvent, error)
}

type Materializer interface {
    Materialize(ctx context.Context, event domain.GeneralizedEvent) error
}

type EventLog interface {
    IsProcessed(ctx context.Context, eventID string) (bool, error)
    MarkProcessed(ctx context.Context, eventID string, eventType string) error
}

type GeographyResolver interface {
    ResolveCEP(ctx context.Context, cep string) (domain.Mesoregion, error)
}
```

## Execution Sequence (ALWAYS this order)
1. **Deserialize** -- JSON bytes to domain event struct
2. **Idempotency** -- check if event already processed (skip if yes)
3. **Anonymize** -- suppress PII fields, generalize quasi-identifiers
4. **Materialize** -- upsert into analytical store (dimensions + facts)
5. **Mark Processed** -- record in event_processing_log (ONLY after materialize succeeds)

## Folder Structure
```
internal/ingestion/
  consumer.go        -- NATS JetStream consumer (Start/Stop lifecycle)
  handler.go         -- PipelineHandler (orchestrates the 5-step flow)
  anonymize.go       -- Anonymizer implementation (calls domain functions)
  materialize.go     -- Materializer implementation (calls store interfaces)
  deserialize.go     -- JSON deserialization for each event type
  events/            -- Event type structs matching AsyncAPI schemas
    patient_created.go
    family_member_added.go
    housing_condition_updated.go
    ...
```

## Rules (non-negotiable)
1. **No business logic** -- if a function decides business state, move it to domain
2. **Dependencies are interfaces** -- never concrete types (pgx, NATS)
3. **Context propagation** -- `ctx context.Context` as first parameter everywhere
4. **Error wrapping** -- `fmt.Errorf("operation: %w", err)` for every error return
5. **ACK after materialization** -- never ACK before the data is persisted
6. **Idempotency** -- every event processed at most once (check before materialize)
7. **No direct store imports** -- only interface types
