package gormq

import (
	"fmt"
	"strings"

	"github.com/assembla/cony"
	"github.com/streadway/amqp"
)

// 封装基于cony 的易用包

//MsgHandle 注册函数
type MsgHandle func(amqp.Delivery)

//Consumer 消费者
type Consumer struct {
	cli *cony.Client
	qe  QueueExchange
	cns *cony.Consumer
}

//Publisher 生产者
type Publisher struct {
	cli *cony.Client
	qe  QueueExchange
	pbl *cony.Publisher
}

//Declarer 定义
type Declarer struct {
	que *cony.Queue   // 队列名称
	exc cony.Exchange // 交换机名称
	bnd cony.Binding  //
}

//QueueExchange 定义队列交换机对象
type QueueExchange struct {
	QueueName    string // 队列名称
	RoutingKey   string // key值
	ExchangeName string // 交换机名称
	ExchangeType string // 交换机类型
}

//Extra 额外的参数
type Extras struct {
	QueueDurable       bool
	QueueAutoDelete    bool
	QueueExclusive     bool
	QueueNoWait        bool
	QueueArgs          map[string]interface{}
	ExchangeDurable    bool
	ExchangeAutoDelete bool
	ExchangeExclusive  bool
	ExchangeNoWait     bool
	ExchangeArgs       map[string]interface{}
	BindArgs           map[string]interface{}
}

// 这里新增了类型，标记这个函数。相关技巧后面介绍
type ExtraFunc func(*Extras)

//RabbitMQ mq对象
type RabbitMQ struct {
	cli  *cony.Client
	addr string // 地址
}

//DefaultOptions 默认配置
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

//New 创建rmq
func New(addr string) *RabbitMQ {
	// Construct new client with the flag url
	// and default backoff policy
	cli := cony.NewClient(
		cony.URL(addr),
		cony.Backoff(cony.DefaultBackoff),
	)

	// Client loop sends out declarations(exchanges, queues, bindings
	// etc) to the AMQP server. It also handles reconnecting.
	go func() {
		for cli.Loop() {
			select {
			case err := <-cli.Errors():
				fmt.Printf("Client error: %v\n", err)
			case blocked := <-cli.Blocking():
				fmt.Printf("Client is blocked %v\n", blocked)
			}
		}
	}()

	rmq := &RabbitMQ{
		cli:  cli,
		addr: addr,
	}
	return rmq
}

//BindQueue 绑定队列（创建队列）
func (r *RabbitMQ) BindQueue(exchangeName, exchangeType, routingKey, queueName string) {
	qe := QueueExchange{
		QueueName:    queueName,
		RoutingKey:   routingKey,
		ExchangeName: exchangeName,
		ExchangeType: exchangeType,
	}
	de := r.Declare(qe)
	if qe.QueueName == "" {
		r.cli.Declare([]cony.Declaration{
			cony.DeclareExchange(de.exc),
		})
	} else {
		r.cli.Declare([]cony.Declaration{
			cony.DeclareQueue(de.que),
			cony.DeclareExchange(de.exc),
			cony.DeclareBinding(de.bnd),
		})
	}
}

//Declare 定义队列或交换机
func (r *RabbitMQ) Declare(qe QueueExchange, exts ...ExtraFunc) Declarer {
	// Declarations
	// 设置额外的参数
	extras := DefaultExtras
	// 遍历进行设置
	for _, extFunc := range exts {
		extFunc(&extras)
	}

	// The queue name will be supplied by the AMQP server
	que := &cony.Queue{
		Name:       strings.Trim(qe.QueueName, " "),
		Durable:    extras.QueueDurable,
		AutoDelete: extras.QueueAutoDelete,
		Exclusive:  extras.QueueExclusive,
		Args:       extras.QueueArgs,
	}
	exc := cony.Exchange{
		Name:       strings.Trim(qe.ExchangeName, " "),
		Kind:       qe.ExchangeType,
		Durable:    extras.ExchangeDurable,
		AutoDelete: extras.ExchangeAutoDelete,
		Args:       extras.ExchangeArgs,
	}
	bnd := cony.Binding{
		Queue:    que,
		Exchange: exc,
		Key:      strings.Trim(qe.RoutingKey, " "),
		Args:     extras.BindArgs,
	}

	return Declarer{
		que: que,
		exc: exc,
		bnd: bnd,
	}

}

//Bind 绑定
func (r *RabbitMQ) Bind(de Declarer) {
	slice := make([]cony.Declaration, 0)
	// 根据名字决定是否定义相关对象
	if de.que.Name != "" {
		slice = append(slice, cony.DeclareQueue(de.que))
	}

	if de.exc.Name != "" {
		slice = append(slice, cony.DeclareExchange(de.exc))
	}

	if de.bnd.Key != "" {
		slice = append(slice, cony.DeclareBinding(de.bnd))
	}

	if len(slice) > 0 {
		r.cli.Declare(slice)
	}
}

//NewConsumer 创建
func (r *RabbitMQ) NewConsumer(qe QueueExchange, exts ...ExtraFunc) *Consumer {
	// Construct new client with the flag url
	// and default backoff policy
	cli := r.cli

	// Declarations
	de := r.Declare(qe, exts...)
	r.Bind(de)

	// Declare and register a consumer
	cns := cony.NewConsumer(
		de.que,
		cony.Qos(1),
	)

	cli.Consume(cns)

	return &Consumer{
		cli: cli,
		qe:  qe,
		cns: cns,
	}
}

//Receive 消费消息
func (c *Consumer) Receive(h MsgHandle) {
	for {
		msg := <-c.cns.Deliveries()
		h(msg)
	}
}

//NewPublisher 创建生产者
func (r *RabbitMQ) NewPublisher(qe QueueExchange, exts ...ExtraFunc) *Publisher {
	// Construct new client with the flag url
	// and default backoff policy
	cli := r.cli

	// Declarations
	de := r.Declare(qe, exts...)
	r.Bind(de)

	pbl := cony.NewPublisher(de.exc.Name, qe.RoutingKey)
	cli.Publish(pbl)

	return &Publisher{
		cli: cli,
		qe:  qe,
		pbl: pbl,
	}
}

//Pub 发布消息
func (p *Publisher) Pub(data []byte) error {
	var err error

	// 发送任务消息
	err = p.pbl.Publish(amqp.Publishing{
		ContentType: "text/plain",
		Body:        data,
	})

	return err
}

//PubRaw 发布消息
func (p *Publisher) PubRaw(raw amqp.Publishing) error {
	// 发送任务消息
	return p.pbl.Publish(raw)
}

// 修改 QueueDurable
func SetQueueDurable(flag bool) ExtraFunc {
	return func(exts *Extras) {
		exts.QueueDurable = flag
	}
}

// 修改 QueueAutoDelete
func SetQueueAutoDelete(flag bool) ExtraFunc {
	return func(exts *Extras) {
		exts.QueueAutoDelete = flag
	}
}

// 修改 QueueAutoDelete
func SetQueueExclusive(flag bool) ExtraFunc {
	return func(exts *Extras) {
		exts.QueueAutoDelete = flag
	}
}

// 修改 QueueNoWait
func SetQueueNoWait(flag bool) ExtraFunc {
	return func(exts *Extras) {
		exts.QueueNoWait = flag
	}
}

// 修改 QueueArgs
func SetQueueArgs(args map[string]interface{}) ExtraFunc {
	return func(exts *Extras) {
		exts.QueueArgs = args
	}
}

// 修改 ExchangeDurable
func SetExchangeDurable(flag bool) ExtraFunc {
	return func(exts *Extras) {
		exts.ExchangeDurable = flag
	}
}

// 修改 ExchangeAutoDelete
func SetExchangeAutoDelete(flag bool) ExtraFunc {
	return func(exts *Extras) {
		exts.ExchangeAutoDelete = flag
	}
}

// 修改 QueueAutoDelete
func SetExchangeExclusive(flag bool) ExtraFunc {
	return func(exts *Extras) {
		exts.ExchangeAutoDelete = flag
	}
}

// 修改 QueueNoWait
func SetExchangeNoWait(flag bool) ExtraFunc {
	return func(exts *Extras) {
		exts.ExchangeNoWait = flag
	}
}

// 修改 ExchangeArgs
func SetExchangeArgs(args map[string]interface{}) ExtraFunc {
	return func(exts *Extras) {
		exts.ExchangeArgs = args
	}
}

// 修改 BindArgs
func SetBindArgs(args map[string]interface{}) ExtraFunc {
	return func(exts *Extras) {
		exts.BindArgs = args
	}
}
