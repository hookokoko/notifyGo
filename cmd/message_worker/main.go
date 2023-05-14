package main

import "notifyGo/internal/engine"

// 在这里应该进行消费者的启动
func main() {
	c := engine.NewConsumer([]string{"127.0.0.1:9092"})
	c.ConsumerGroup("message_common", "g_test", "test")
}
