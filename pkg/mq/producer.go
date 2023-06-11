package mq

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

type Producer struct {
	hosts                []string
	topicBalanceMappings map[string]Balancer
}

func NewProducer(mqCfg *Config) *Producer {
	var tbMappings map[string]Balancer
	tbMappings = make(map[string]Balancer, len(mqCfg.TopicMappings))

	//	为每一个channel创建好balancer
	for channel, topics := range mqCfg.TopicMappings {
		bala := NewBalanceBuilder(channel, topics.Topics).Build(topics.Strategy)
		tbMappings[channel] = bala
	}

	go func() {
		for {
			mu.Lock()
			changeSignal.Wait()
			log.Println("producer update topics")
			for channel, topics := range mqCfg.TopicMappings {
				bala := NewBalanceBuilder(channel, topics.Topics).Build(topics.Strategy)
				tbMappings[channel] = bala
			}
			mu.Unlock()
		}
	}()

	return &Producer{
		hosts:                mqCfg.Host,
		topicBalanceMappings: tbMappings,
	}
}

// Send data里面应该包含channel信息，方法找到topic
func (p *Producer) Send(channel string, data []byte) {
	config := sarama.NewConfig()
	// 异步生产者不建议把 Errors 和 Successes 都开启，一般开启 Errors 就行
	// 同步生产者就必须都开启，因为会同步返回发送成功或者失败
	config.Producer.Return.Errors = true    // 设定是否需要返回错误信息
	config.Producer.Return.Successes = true // 设定是否需要返回成功信息
	producer, err := sarama.NewAsyncProducer(p.hosts, config)
	if err != nil {
		log.Fatal("NewSyncProducer err:", err)
	}
	var (
		wg                                   sync.WaitGroup
		enqueued, timeout, successes, errors int
	)
	// [!important] 异步生产者发送后必须把返回值从 Errors 或者 Successes 中读出来 不然会阻塞 sarama 内部处理逻辑 导致只能发出去一条消息
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range producer.Successes() {
			// log.Printf("[Producer] Success: key:%v msg:%+v \n", s.Key, s.Value)
			successes++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range producer.Errors() {
			log.Printf("[Producer] Errors：err:%v msg:%+v \n", e.Msg, e.Err)
			errors++
		}
	}()

	// 根据channel类型，和路由策略选取发送的topic
	topic, err := p.topicBalanceMappings[channel].GetNext()
	if err != nil {
		log.Printf("[Producer] choose topic fail: channel:%v error:%+v \n", channel, err)
	}
	msg := &sarama.ProducerMessage{Topic: topic, Key: nil, Value: sarama.ByteEncoder(data)}
	// 异步发送只是写入内存了就返回了，并没有真正发送出去
	// sarama 库中用的是一个 channel 来接收，后台 goroutine 异步从该 channel 中取出消息并真正发送
	// select + ctx 做超时控制,防止阻塞 producer.Input() <- msg 也可能会阻塞
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	select {
	case producer.Input() <- msg:
		enqueued++
	case <-ctx.Done():
		timeout++
	}
	cancel()

	producer.AsyncClose()
	wg.Wait()
	log.Printf("发送完毕[%s] enqueued:%d timeout:%d successes: %d errors: %d\n", topic, enqueued,
		timeout, successes, errors)
}
