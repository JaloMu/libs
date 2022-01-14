package consumer

import (
	"context"
	"strings"

	"github.com/apache/rocketmq-client-go/v2"

	"go.uber.org/zap"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func NewPush(cs ConsumeStruct, fn func(msg *primitive.MessageExt) (err error)) (err error) {
	traceCfg := &primitive.TraceConfig{
		Access:       primitive.Local,
		NamesrvAddrs: cs.Address,
	}
	address := primitive.NewPassthroughResolver(cs.Address)
	pushDo, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(cs.GroupName),
		consumer.WithNsResolver(address),
		consumer.WithRetry(cs.Retry),
		consumer.WithTrace(traceCfg),
	)

	if err != nil {
		return
	}
	selector := consumer.MessageSelector{}
	if len(cs.Tags) != 0 {
		selector.Type = consumer.TAG
		selector.Expression = strings.Join(cs.Tags, " || ")
	}

	err = pushDo.Subscribe(cs.Topic, selector, func(ctx context.Context, messages ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range messages {
			err = fn(msg)
			// 消息处理失败 打印失败日志
			if err != nil {
				zap.L().Named("rocketmq").Error("_rocker_push_consume_err",
					zap.Error(err),
					zap.String("topic", cs.Topic),
					zap.Any("message", msg),
				)
				continue
			}
			zap.L().Named("rocketmq").Info("_rocker_push_consume_out",
				zap.String("topic", cs.Topic),
				zap.Any("message", msg),
			)
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		return
	}
	err = pushDo.Start()
	return
}
