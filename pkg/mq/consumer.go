package mq

import (
	"context"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

// 这里写consumer的抽象方法，然后具体的消费逻辑写到engine.go里面，然后路由到sender中不同的处理方法里面
/*
	消费者在初始化的时候，应该加载一些配置文件，然后就是创建一个线程池，
	消费的时候把对应的处理函数放到线程池中
*/

type ConsumerGroup struct {
	handler sarama.ConsumerGroupHandler
	sCg     sarama.ConsumerGroup
	topics  []string
}

// TODO handler的动态加载
func NewConsumerGroup(mqCfg *Config, channel string, handler sarama.ConsumerGroupHandler) *ConsumerGroup {
	sCfg := sarama.NewConfig()
	sCfg.Consumer.Return.Errors = true

	topics := mqCfg.GetTopicsByChannel(channel)
	groupId := mqCfg.GetGroupIdByChannel(channel)

	// 这里面一个channel就是统一的一个group，就没法更细粒度的区分验证码消息、营销类消息这种了，
	// 但是似乎现在看区分也没什么必要了，因为即使我消息阻塞了可以扩topic，
	// 况且无论是哪种消息，只要channel一样，处理逻辑就是一样的

	cg, err := sarama.NewConsumerGroup(mqCfg.Host, groupId, sCfg)
	if err != nil {
		log.Fatal("NewConsumerGroup err: ", err)
	}

	return &ConsumerGroup{
		handler: handler,
		sCg:     cg,
		topics:  topics,
	}
}

func (c *ConsumerGroup) Start(ctx context.Context) {
	defer func() { _ = c.sCg.Close() }()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		fmt.Println("running: ", c.topics)
		/*
			![important]
			应该在一个无限循环中不停地调用 Consume()
			因为每次 Rebalance 后需要再次执行 Consume() 来恢复连接
			Consume 开始才发起 Join Group 请求 如果当前消费者加入后成为了 消费者组 leader,则还会进行 Rebalance 过程，从新分配
			组内每个消费组需要消费的 topic 和 partition，最后 Sync Group 后才开始消费
			具体信息见 https://github.com/lixd/kafka-go-example/issues/4
		*/
		err := c.sCg.Consume(ctx, c.topics, c.handler)
		if err != nil {
			log.Println("Consume err: ", err)
		}
	}
}
