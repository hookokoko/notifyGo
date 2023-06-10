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
	hm := sender.NewHandlerManager()
	emailPool := worker.NewPoolExecutor()
	smsPool := worker.NewPoolExecutor()
	pushPool := worker.NewPoolExecutor()
	mqCfg := mq.NewConfig("/Users/hooko/GolandProjects/notifyGo/config/kafka_topic.toml")
	emailHandler, _ := hm.Get("email")
	smsHandler, _ := hm.Get("sms")
	pushHandler, _ := hm.Get("push")
	startEmail(emailHandler, emailPool, mqCfg)
	startSMS(smsHandler, smsPool, mqCfg)
	startPush(pushHandler, pushPool, mqCfg)
	select {}
}

func startEmail(h sender.IHandler, pool *worker.PoolExecutor, cfg mq.Config) {
	handler := worker.NewConsumerHandler(h, pool)
	cg1 := mq.NewConsumerGroup(cfg, "email.verification", []string{cfg.Topics.EmailMappings.Verification}, handler)
	cg2 := mq.NewConsumerGroup(cfg, "email.notification", []string{cfg.Topics.EmailMappings.Notification}, handler)
	cg3 := mq.NewConsumerGroup(cfg, "email.marketing", []string{cfg.Topics.EmailMappings.Marketing}, handler)
	go cg1.Start(context.TODO())
	go cg2.Start(context.TODO())
	go cg3.Start(context.TODO())
}

func startSMS(h sender.IHandler, pool *worker.PoolExecutor, cfg mq.Config) {
	handler := worker.NewConsumerHandler(h, pool)
	cg1 := mq.NewConsumerGroup(cfg, "sms.verification", []string{cfg.Topics.SmsMappings.Verification}, handler)
	cg2 := mq.NewConsumerGroup(cfg, "sms.notification", []string{cfg.Topics.SmsMappings.Notification}, handler)
	cg3 := mq.NewConsumerGroup(cfg, "sms.marketing", []string{cfg.Topics.SmsMappings.Marketing}, handler)
	go cg1.Start(context.TODO())
	go cg2.Start(context.TODO())
	go cg3.Start(context.TODO())
}

func startPush(h sender.IHandler, pool *worker.PoolExecutor, cfg mq.Config) {
	handler := worker.NewConsumerHandler(h, pool)
	cg1 := mq.NewConsumerGroup(cfg, "push.verification", []string{cfg.Topics.PushMappings.Verification}, handler)
	cg2 := mq.NewConsumerGroup(cfg, "push.notification", []string{cfg.Topics.PushMappings.Notification}, handler)
	cg3 := mq.NewConsumerGroup(cfg, "push.marketing", []string{cfg.Topics.PushMappings.Marketing}, handler)
	go cg1.Start(context.TODO())
	go cg2.Start(context.TODO())
	go cg3.Start(context.TODO())
}
