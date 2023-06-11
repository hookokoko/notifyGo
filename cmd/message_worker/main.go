package main

import (
	"context"
	"notifyGo/internal/worker"
	"notifyGo/internal/worker/sender"
	"notifyGo/pkg/mq"
)

// 消费者拓扑图，原则就是所有的topic都是等价的，也就是消费者的消费逻辑不绑定topic
// 每种渠道(短信、邮件、push)创建多个topic，同一种渠道的topic是等价的，也就是说发送消息的时候按照一定的负载均衡策略去选择某一个topic；
// 同一个渠道的消费者逻辑是同一个；
// 其中，每调用一次NewConsumerGroup()，就会分别为同一个channel的3个topic的每个topic创建1个消费者；
// topic可以无限加，防止某一个channel的topic出现阻塞。
// topic的动态加载待支持
// sms
// 	- topic-high
//		- consumer_group_sms_high
// 	- topic-medium
// 		- consumer_group_sms_medium
//	- topic-low
// 		- consumer_group_sms_low
// email
// 	- topic-high
// 	- topic-medium
//	- topic-low
// push
// 	- topic-high
// 	- topic-medium
//	- topic-low

// 在这里应该进行消费者的启动
func main() {
	mqCfg := mq.NewConfig("/Users/hooko/GolandProjects/notifyGo/config/kafka_topic.toml")

	hm := sender.NewHandlerManager()
	emailPool := worker.NewPoolExecutor()
	smsPool := worker.NewPoolExecutor()
	pushPool := worker.NewPoolExecutor()

	emailHandler, _ := hm.Get("email")
	smsHandler, _ := hm.Get("sms")
	pushHandler, _ := hm.Get("push")

	startEmail(emailHandler, emailPool, mqCfg)
	startSMS(smsHandler, smsPool, mqCfg)
	startPush(pushHandler, pushPool, mqCfg)

	select {}
}

func startEmail(h sender.IHandler, pool *worker.PoolExecutor, cfg *mq.Config) {
	handler := worker.NewConsumerHandler(h, pool)
	cg1 := mq.NewConsumerGroup(cfg, "email", handler)
	go cg1.Start(context.TODO())
}

func startSMS(h sender.IHandler, pool *worker.PoolExecutor, cfg *mq.Config) {
	handler := worker.NewConsumerHandler(h, pool)
	cg1 := mq.NewConsumerGroup(cfg, "sms", handler)
	go cg1.Start(context.TODO())
}

func startPush(h sender.IHandler, pool *worker.PoolExecutor, cfg *mq.Config) {
	handler := worker.NewConsumerHandler(h, pool)
	cg1 := mq.NewConsumerGroup(cfg, "push", handler)
	go cg1.Start(context.TODO())
}
