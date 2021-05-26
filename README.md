# gormq
golang rabbitmq

## 使用方式

```
# 创建rabbitmq连接
addr := "amqp://admin:admin@127.0.0.1:5672"
rabbitmq := gormq.New(addr)

# 队列和交换机 创建
queueExchange := gormq.QueueExchange{
    QueueName:    "myname11",
    RoutingKey:   "iot.report.opendoor.insert",
    ExchangeName: "iot.opendoor",
    ExchangeType: "topic",
}

# 额外参数的默认值为
var DefaultExtras = Extras{
	QueueDurable:       true,
	QueueAutoDelete:    false,
	QueueExclusive:     false,
	QueueNoWait:        true,
	QueueArgs:          nil,
	ExchangeDurable:    true,
	ExchangeAutoDelete: false,
	ExchangeExclusive:  false,
	ExchangeNoWait:     true,
	ExchangeArgs:       nil,
	BindArgs:           nil,
}

# 设置 额外的参数 配置(每一项都是可选的)
extras := []gormq.ExtraFunc{
    gormq.SetQueueDurable(true),
    gormq.SetQueueAutoDelete(false),
    gormq.SetQueueExclusive(false),
    gormq.SetQueueNoWait(true),
    gormq.SetQueueArgs(map[string]interface{}{
        "x-message-ttl": 30000,
    }),

    gormq.SetExchangeDurable(true),
    gormq.SetExchangeAutoDelete(false),
    gormq.SetExchangeExclusive(false),
    gormq.SetExchangeNoWait(true),
    gormq.SetExchangeArgs(map[string]interface{}{
        
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

# 发送原始数据
publisher.PubRaw(amqp.Publishing{
    ContentType: "text/plain",
    Body:        []byte("dhhs"),
})
```

## todo

- err 返回打印
