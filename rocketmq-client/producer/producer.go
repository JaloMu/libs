package producer

import (
	"context"
	"strings"

	"github.com/JaloMu/libs/trace"
	"go.uber.org/zap"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type ProduceStruct struct {
	Address []string `json:"address"`
	Retry   int      `json:"retry"`
}

var producerPublic rocketmq.Producer

func NewProducer(prod ProduceStruct) (err error) {
	traceCfg := &primitive.TraceConfig{
		Access:   primitive.Local,
		Resolver: primitive.NewPassthroughResolver(prod.Address),
	}
	producerPublic, err = rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(prod.Address)),
		producer.WithRetry(prod.Retry),
		producer.WithTrace(traceCfg),
	)
	if err != nil {
		return err
	}
	if err = producerPublic.Start(); err != nil {
		return err
	}
	return nil
}

type Message struct {
	Topic string   `json:"topic"`
	Msg   string   `json:"msg"`
	Tags  []string `json:"tags"`
}

func SendMQ(ctx context.Context, traceKey string, message Message) (err error) {
	tr := ctx.Value(traceKey)
	traceContext, ok := tr.(*trace.TraceContext)
	if !ok {
		traceContext = &trace.TraceContext{}
	}
	msg := primitive.NewMessage(message.Topic, []byte(message.Msg))
	zap.L().Named("rocketmq").Info("_rocker_pull_produce_in",
		zap.String("traceId", traceContext.TraceId),
		zap.String("child_spanId", traceContext.CSpanId),
		zap.String("spanId", traceContext.SpanId),
		zap.String("topic", message.Topic),
		zap.String("tags", strings.Join(message.Tags, ", ")),
		zap.String("message", message.Msg),
	)
	for _, tag := range message.Tags {
		msg.WithTag(tag)
		res, err := producerPublic.SendSync(ctx, msg)
		if err != nil {
			zap.L().Named("rocketmq").Error("_rocker_pull_produce_err",
				zap.Any("error", err),
				zap.String("traceId", traceContext.TraceId),
				zap.String("child_spanId", traceContext.CSpanId),
				zap.String("spanId", traceContext.SpanId),
				zap.String("topic", message.Topic),
				zap.String("tags", strings.Join(message.Tags, ", ")),
				zap.String("message", message.Msg),
				zap.Any("result", res),
			)
			continue
		}
		zap.L().Named("rocketmq").Info("_rocker_pull_produce_out",
			zap.String("traceId", traceContext.TraceId),
			zap.String("child_spanId", traceContext.CSpanId),
			zap.String("spanId", traceContext.SpanId),
			zap.String("topic", message.Topic),
			zap.String("tags", strings.Join(message.Tags, ", ")),
			zap.String("message", message.Msg),
			zap.Any("result", res),
		)
	}
	return nil
}

func Close() (err error) {
	return producerPublic.Shutdown()
}
