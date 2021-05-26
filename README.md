# gormq
golang rabbitmq

## 使用方式

```
# 创建rabbitmq连接
addr := "amqp://admin:admin@127.0.0.1:5672"
rmq := gormq.New(addr)

# 队列和交换机 创建
queueExchange := gormq.QueueExchange{
    QueueName:    "myname11",
    RoutingKey:   "iot.report.opendoor.insert",
    ExchangeName: "iot.opendoor",
    ExchangeType: "topic",
}

# 额外的参数 配置
extras := []gormq.ExtraFunc{
    gormq.SetQueueDurable(false),
    gormq.SetQueueAutoDelete(false),
    gormq.SetQueueArgs(map[string]interface{}{
        "x-message-ttl": 30000,
    }),
}

# 创建消费者
consumer := rabbitmq.NewConsumer(queueExchange, extras...)

# 消费数据
consumer.Receive(func(msg amqp.Delivery) {
    msg.Ack(false)
    fmt.Println(string(msg.Body))
})

# 创建生产者
publisher := rabbitmq.NewPublisher(queueExchange, extras...)

# 发送消息
err := publisher.Pub([]byte("hhshs"))
fmt.Println(err)
```
