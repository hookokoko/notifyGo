package main

import (
	"context"
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
	test1()
}

func normal() {
	hm := sender.NewHandlerManager()
	emailPool := worker.NewPoolExecutor()
	emailHandler := worker.NewConsumerHandler(hm, emailPool)

	mqCfg := mq.NewConfig("/Users/tiger/GolandProjects/notifyGo/config/kafka_topic.toml")

	// groupId 标识起动的消费者属于对应哪个消费者组
	cg1 := mq.NewConsumerGroup(mqCfg, "email.marketing", []string{mqCfg.Topics.EmailMappings.Marketing}, emailHandler)
	go cg1.Start(context.TODO())

	cg2 := mq.NewConsumerGroup(mqCfg, "email.notification", []string{mqCfg.Topics.EmailMappings.Notification}, emailHandler)
	go cg2.Start(context.TODO())

	cg3 := mq.NewConsumerGroup(mqCfg, "email.verification", []string{mqCfg.Topics.EmailMappings.Verification}, emailHandler)
	go cg3.Start(context.TODO())

	select {}
}

// 单个topic、消费者组，多个消费者
func test1() {
	hm := sender.NewHandlerManager()
	emailPool := worker.NewPoolExecutor()
	emailHandler := worker.NewConsumerHandler(hm, emailPool)

	mqCfg := mq.NewConfig("/Users/tiger/GolandProjects/notifyGo/config/kafka_topic.toml")

	ctx := context.Background()
	// groupId 标识起动的消费者属于对应哪个消费者组
	cg1 := mq.NewConsumerGroup(mqCfg, "email.marketing", []string{mqCfg.Topics.EmailMappings.Marketing}, emailHandler)
	cg2 := mq.NewConsumerGroup(mqCfg, "email.marketing", []string{mqCfg.Topics.EmailMappings.Marketing}, emailHandler)
	cg3 := mq.NewConsumerGroup(mqCfg, "email.marketing", []string{mqCfg.Topics.EmailMappings.Marketing}, emailHandler)
	go cg1.Start(ctx)
	go cg2.Start(ctx)
	go cg3.Start(ctx)
	select {}
}

func startEmail(hm *sender.HandleManager, pool *worker.PoolExecutor, cfg mq.Config) {
	mqCfg := mq.NewConfig("/Users/hooko/GolandProjects/notifyGo/config/kafka_topic.toml")
	handler := worker.NewConsumerHandler(hm, pool)
	cg := mq.NewConsumerGroup(cfg, "email.notice", []string{mqCfg.Topics.EmailMappings.Notification}, handler)
	go cg.Start(context.TODO())
}

//func startSMS(hm *sender.HandleManager, pool *worker.PoolExecutor, cfg mq.Config) {
//	handler := worker.NewConsumerHandler(hm, pool)
//	cg1 := mq.NewConsumerGroup(cfg, "sms.notice", handler)
//	cg1.Start([]string{"notify_go_common"})
//
//	cg2 := mq.NewConsumerGroup(cfg, "sms.market", handler)
//	cg2.Start([]string{"notify_go_common"})
//}
