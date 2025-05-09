package appconfig

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/bridges/otellogrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/log/global"
	meter "go.opentelemetry.io/otel/metric"
	otelLog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type ApplicationMetrics struct {
	httpRequestsReceivedTotal  meter.Int64Counter
	httpRequestDurationSeconds meter.Float64Histogram
}

const (
	APPLICATION_NAME = "deployvia"
	SYSTEM_NAME      = "core"
)

func ConfigureOpenTelemetry(ctx context.Context) (*ApplicationMetrics, error) {
	resource, err := resource.New(
		ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String(APPLICATION_NAME)),
		resource.WithAttributes(semconv.ServiceNamespaceKey.String(SYSTEM_NAME)),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		log.Errorf("Failed to create resource: %v", err)

		return nil, errors.New("FailedToCreateResource")
	}

	if err := configureLogs(ctx, resource); err != nil {
		log.Errorf("Failed to configure logs: %v", err)

		return nil, errors.New("FailedToConfigureLogs")
	}

	applicationMetrics, err := configureMetrics(ctx, resource)
	if err != nil {
		log.Errorf("Failed to configure metrics: %v", err)

		return nil, errors.New("FailedToConfigureMetrics")
	}

	return applicationMetrics, nil
}

func configureLogs(ctx context.Context, resource *resource.Resource) error {
	logExporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return err
	}

	processor := otelLog.NewBatchProcessor(logExporter)
	loggerProvider := otelLog.NewLoggerProvider(
		otelLog.WithResource(resource),
		otelLog.WithProcessor(processor),
	)

	global.SetLoggerProvider(loggerProvider)

	hook := otellogrus.NewHook(SYSTEM_NAME, otellogrus.WithLoggerProvider(loggerProvider))
	log.AddHook(hook)

	return nil
}

func configureMetrics(ctx context.Context, resource *resource.Resource) (*ApplicationMetrics, error) {
	metricExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus exporter: %v", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
		metric.WithResource(resource),
	)
	otel.SetMeterProvider(meterProvider)

	metrics := meterProvider.Meter(APPLICATION_NAME)

	httpRequestsReceivedTotal, err := metrics.Int64Counter(
		"http_requests_received_total",
		meter.WithDescription("Total number of HTTP requests received"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create counter: %s", err)
	}

	httpRequestDurationSeconds, err := metrics.Float64Histogram(
		"http_request_duration_seconds",
		meter.WithDescription("The duration of HTTP requests processed by Gin, in seconds."),
		meter.WithExplicitBucketBoundaries(
			0.001,
			0.002,
			0.005,
			0.01,
			0.02,
			0.05,
			0.1,
			0.2,
			0.5,
			1,
			2,
			5,
			10,
			20,
			60,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create histogram: %s", err)
	}

	return &ApplicationMetrics{
		httpRequestsReceivedTotal:  httpRequestsReceivedTotal,
		httpRequestDurationSeconds: httpRequestDurationSeconds,
	}, nil
}

func Metrics(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()

		latency := time.Since(t)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		endpoint := c.Request.URL.Path

		meterAttributes := []attribute.KeyValue{
			attribute.Key("code").Int(statusCode),
			attribute.Key("method").String(method),
			attribute.Key("endpoint").String(endpoint),
			// TODO: this shouldn't be needed, we don't have controllers
			attribute.Key("controller").String("gin"),
		}

		config.ApplicationMetrics.httpRequestDurationSeconds.Record(
			c.Request.Context(),
			latency.Seconds(),
			meter.WithAttributes(meterAttributes...),
		)

		config.ApplicationMetrics.httpRequestsReceivedTotal.Add(
			c.Request.Context(),
			1,
			meter.WithAttributes(meterAttributes...),
		)
	}
}
