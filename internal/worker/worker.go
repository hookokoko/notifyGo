package worker

import (
	"context"
	"encoding/json"
	"log"
	"notifyGo/internal"
	"notifyGo/internal/worker/sender"

	"github.com/Shopify/sarama"
)

func NewConsumerHandler(hm *sender.HandleManager, pool *PoolExecutor) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		handlers: hm,
		pool:     pool,
	}
}

// ConsumerGroupHandler 实现 sarama.ConsumerGroup 接口，作为自定义ConsumerGroup
type ConsumerGroupHandler struct {
	name     string
	handlers *sender.HandleManager
	pool     *PoolExecutor
}

// ConsumeClaim 具体的消费逻辑
func (h *ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("[consumer] name:%s topic:%q partition:%d offset:%d\n", h.name, msg.Topic, msg.Partition, msg.Offset)

		// 1. 这里将msg序列化，获取消息类型
		//    其中，task.ITarget是接口类型，该如何序列化？
		taskInfo := new(internal.Task)
		err := json.Unmarshal(msg.Value, taskInfo)
		if err != nil {
			log.Printf("[consumer] name:%s topic:%q partition:%d offset:%d <ERROR: 解析消息体出错 %v>\n",
				h.name, msg.Topic, msg.Partition, msg.Offset, err)
		}

		// 2. 根据消息渠道找到对应的handler
		hAction, err := h.handlers.Get(taskInfo.SendChannel)
		if err != nil {
			log.Printf("[consumer] name:%s topic:%q partition:%d offset:%d <ERROR: 找不到发送消息的handler %s>\n",
				h.name, msg.Topic, msg.Partition, msg.Offset, taskInfo.SendChannel)
			return err
		}

		// 3. 通过TaskInfo和具体的执行handler，构造待执行的Task, 然后提交对应的协程池
		err = h.pool.Submit(context.TODO(), NewTask(taskInfo, hAction))
		if err != nil {
			log.Printf("[consumer] name:%s topic:%q partition:%d offset:%d <ERROR: 执行消息发送失败 %v>\n",
				h.name, msg.Topic, msg.Partition, msg.Offset, err)
			return err
		}

		// 标记消息已被消费 内部会更新 consumer offset
		sess.MarkMessage(msg, "")
	}
	return nil
}

// Setup 执行在 获得新 session 后 的第一步, 在 ConsumeClaim() 之前
func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup 执行在 session 结束前, 当所有 ConsumeClaim goroutines 都退出时
func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
