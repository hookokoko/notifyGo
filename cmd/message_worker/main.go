package main

import (
	"notifyGo/internal/worker"
	"notifyGo/internal/worker/sender"
	"notifyGo/pkg/mq"
)

// 消费者拓扑图，原则就是所有的topic都是等价的，也就是消费者的消费逻辑不绑定topic
// 以high这个topic为例
// high
// - cg_email_验
// 		consumer_1
// 		consumer_2
// 		consumer_3
// 		...
// - cg_sms_验
// 		consumer_1
// 		consumer_2
// 		consumer_3
// 		...
// - cg_push_验
// 		consumer_1
// 		consumer_2
// 		consumer_3
// 		...

// 在这里应该进行消费者的启动
func main() {
	//c := engine.NewConsumer([]string{"127.0.0.1:9092"})
	//c.ConsumerGroup("message_common", "g_test", "test")

	hm := sender.NewHandlerManager()
	emailPool := worker.NewPoolExecutor()
	emailHandler := worker.NewConsumerHandler(hm, emailPool)

	mqCfg := mq.NewConfig("../../config/kafka_topic.toml")

	// groupId 标识起动的消费者属于对应哪个消费者组
	cg := mq.NewConsumerGroup(mqCfg, "email.marketing", emailHandler)

	// 这里面表示消费哪些topic消息
	// 通过配置文件查询是LowTopic
	cg.Start([]string{mqCfg.Topics.EmailMappings.Marketing})
}

func startEmail(hm *sender.HandleManager, pool *worker.PoolExecutor, cfg mq.Config) {
	handler := worker.NewConsumerHandler(hm, pool)
	cg1 := mq.NewConsumerGroup(cfg, "email.notice", handler)
	cg1.Start([]string{"notify_go_common"})

	cg2 := mq.NewConsumerGroup(cfg, "email.market", handler)
	cg2.Start([]string{"notify_go_common"})
}

func startSMS(hm *sender.HandleManager, pool *worker.PoolExecutor, cfg mq.Config) {
	handler := worker.NewConsumerHandler(hm, pool)
	cg1 := mq.NewConsumerGroup(cfg, "sms.notice", handler)
	cg1.Start([]string{"notify_go_common"})

	cg2 := mq.NewConsumerGroup(cfg, "sms.market", handler)
	cg2.Start([]string{"notify_go_common"})
}
