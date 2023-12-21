package kafka

import (
	"context"
	"github.com/sefikcan/read-time-trade/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type Producer interface {
	PublishMessage(ctx context.Context, kafkaMessages ...kafka.Message) error
	Close() error
}

type producer struct {
	log     logger.Logger
	brokers []string
	w       *kafka.Writer
}

func NewProducer(log logger.Logger, brokers []string) *producer {
	return &producer{
		log:     log,
		brokers: brokers,
		w:       NewWriter(brokers, kafka.LoggerFunc(log.Errorf)),
	}
}

func (p *producer) PublishMessage(ctx context.Context, kafkaMessages ...kafka.Message) error {
	return p.w.WriteMessages(ctx, kafkaMessages...)
}

func (p *producer) Close() error {
	return p.w.Close()
}
