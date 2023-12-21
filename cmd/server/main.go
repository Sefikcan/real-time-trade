package main

import (
	"github.com/opentracing/opentracing-go"
	"github.com/sefikcan/read-time-trade/internal/server"
	"github.com/sefikcan/read-time-trade/pkg/config"
	"github.com/sefikcan/read-time-trade/pkg/logger"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	jaegerLog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"io"
	"log"
)

func main() {
	log.Println("Starting api server")

	cfg := config.NewConfig()

	zapLogger := logger.NewLogger(cfg)
	zapLogger.InitLogger()
	zapLogger.Infof("AppVersion: %s, LogLevel: %s, Mode: %s, SSL: %v", cfg.Server.AppVersion, cfg.Logger.Level, cfg.Server.Mode, false)

	jaegerConfigInstance := jaegerCfg.Configuration{
		ServiceName: cfg.Metric.ServiceName,
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerCfg.ReporterConfig{
			LogSpans:           cfg.Jaeger.LogSpans,
			LocalAgentHostPort: cfg.Jaeger.Host,
		},
	}

	tracer, closer, err := jaegerConfigInstance.NewTracer(
		jaegerCfg.Logger(jaegerLog.StdLogger),
		jaegerCfg.Metrics(metrics.NullFactory),
	)
	if err != nil {
		log.Fatal("can't create tracer", err)
	}
	zapLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {

		}
	}(closer)
	zapLogger.Info("Opentracing connected")
	
	s := server.NewServer(cfg, zapLogger)
	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}
