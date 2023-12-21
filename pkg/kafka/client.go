package kafka

import (
	"context"
	"github.com/sefikcan/read-time-trade/pkg/config"
	"github.com/segmentio/kafka-go"
)

func NewKafkaConn(ctx context.Context, cfg *config.Config) (*kafka.Conn, error) {
	return kafka.DialContext(ctx, "tcp", cfg.Kafka.Brokers[0])
}
