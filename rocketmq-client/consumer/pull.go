package consumer

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

type ConsumeStruct struct {
	Retry     int      `json:"retry"`
	GroupName string   `json:"group_name"`
	Address   []string `json:"address"`
	Topic     string   `json:"topic"`
	TraceKey  string   `json:"trace_key"`
	Tags      []string `json:"tags"`
}

func NewPullConsume(cs ConsumeStruct, fn func(*primitive.MessageExt) error) (err error) {
	traceCfg := &primitive.TraceConfig{
		Access:       primitive.Local,
		NamesrvAddrs: cs.Address,
	}
	address, err := primitive.NewNamesrvAddr(cs.Address...)
	if err != nil {
		return
	}
	do, err := consumer.NewPullConsumer(
		consumer.WithGroupName(cs.GroupName),
		consumer.WithNameServer(address),
		consumer.WithRetry(cs.Retry),
		consumer.WithTrace(traceCfg),
	)
	if err != nil {
		return
	}
	err = do.Start()
	if err != nil {
		return
	}
	defer do.Shutdown()
	selector := consumer.MessageSelector{}
	if len(cs.Tags) != 0 {
		selector.Type = consumer.TAG
		selector.Expression = strings.Join(cs.Tags, " || ")
	}
	//ctx context.Context, topic string, selector MessageSelector, numbers int
	for {
		resp, err := do.Pull(context.Background(), cs.Topic, selector, 1)
		if err != nil {
			zap.L().Named("rocketmq").Error("_rocker_pull_consume_err",
				zap.Error(err),
				zap.String("topic", cs.Topic),
				zap.Any("pull_result", resp),
			)
			continue
		}
		if resp.Status == primitive.PullFound {
			zap.L().Named("rocketmq").Info("_rocker_pull_consume_in",
				zap.String("topic", cs.Topic),
				zap.Any("pull_result", resp),
			)
			for _, msg := range resp.GetMessageExts() {
				err = fn(msg)
				if err != nil {
					zap.L().Named("rocketmq").Error("_rocker_pull_consume_err",
						zap.Error(err),
						zap.String("topic", cs.Topic),
						zap.Any("message", msg),
					)
					continue
				}
				zap.L().Named("rocketmq").Info("_rocker_pull_consume_out",
					zap.String("topic", cs.Topic),
					zap.Any("message", msg),
				)
			}
		}
	}
}
