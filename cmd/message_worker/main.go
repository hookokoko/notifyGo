package main

import (
	"notifyGo/internal/worker"
	"notifyGo/internal/worker/sender"
	"notifyGo/pkg/mq"
)

// 在这里应该进行消费者的启动
func main() {
	//c := engine.NewConsumer([]string{"127.0.0.1:9092"})
	//c.ConsumerGroup("message_common", "g_test", "test")

	hm := sender.NewHandlerManager()
	emailPool := worker.NewPoolExecutor()
	emailHandler := worker.NewConsumerHandler(hm, emailPool)

	cfg := mq.NewConfig([]string{"127.0.0.1:9092"})

	// groupId 标识起动的消费者属于对应哪个消费者组
	cg := mq.NewConsumerGroup(cfg, "email", emailHandler)

	// 这里面表示消费哪些topic消息
	cg.Start([]string{"email"})
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
