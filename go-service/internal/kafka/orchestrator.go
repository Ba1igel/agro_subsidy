package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"

	"agro-subsidy/go-service/internal/config"
	"agro-subsidy/go-service/internal/model"
	"agro-subsidy/go-service/internal/worker"
)

// Orchestrator owns the Kafka consumer group and result producer.
// It feeds tasks into the worker pool and publishes scored results back to Kafka.
type Orchestrator struct {
	reader *kafka.Reader
	writer *kafka.Writer
	pool   *worker.Pool
}

func NewOrchestrator(cfg config.Config, pool *worker.Pool) *Orchestrator {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBrokers},
		Topic:    cfg.KafkaInputTopic,
		GroupID:  cfg.KafkaGroupID,
		MinBytes: 1,
		MaxBytes: 10e6, // 10 MB
	})

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokers),
		Topic:                  cfg.KafkaOutputTopic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	return &Orchestrator{
		reader: reader,
		writer: writer,
		pool:   pool,
	}
}

// Run is the main loop: consume → enqueue → (goroutine) publish.
// Returns when ctx is cancelled.
func (o *Orchestrator) Run(ctx context.Context) error {
	go o.publishResults(ctx)

	for {
		msg, err := o.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // clean shutdown
			}
			log.Printf("[kafka] read error: %v", err)
			continue
		}

		var task model.SubsidiesTask
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			log.Printf("[kafka] unmarshal error offset=%d: %v", msg.Offset, err)
			continue
		}

		if !o.pool.Submit(ctx, task) {
			return nil // ctx cancelled
		}
	}
}

func (o *Orchestrator) publishResults(ctx context.Context) {
	for {
		select {
		case result, ok := <-o.pool.Results():
			if !ok {
				return
			}
			body, err := json.Marshal(result)
			if err != nil {
				log.Printf("[kafka] marshal result error task=%s: %v", result.TaskID, err)
				continue
			}
			if err := o.writer.WriteMessages(ctx, kafka.Message{
				Key:   []byte(result.TaskID),
				Value: body,
			}); err != nil && ctx.Err() == nil {
				log.Printf("[kafka] write result error task=%s: %v", result.TaskID, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (o *Orchestrator) Close() {
	if err := o.reader.Close(); err != nil {
		log.Printf("[kafka] reader close: %v", err)
	}
	if err := o.writer.Close(); err != nil {
		log.Printf("[kafka] writer close: %v", err)
	}
}
