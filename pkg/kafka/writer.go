package kafka

import (
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

func NewWriter(brokers []string, errLogger kafka.Logger) *kafka.Writer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           writerRequiredAcks,
		MaxAttempts:            writerMaxAttempts,
		ErrorLogger:            errLogger,
		Compression:            compress.Snappy,
		ReadTimeout:            writerReadTimeout,
		WriteTimeout:           writerWriteTimeout,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}
	return w
}
