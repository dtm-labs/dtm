package dtmsvr

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/yedf/dtm/common"
)

type Rabbitmq struct {
	Config      RabbitmqConfig
	ChannelPool *sync.Pool
}

type RabbitmqConfig struct {
	Host          string
	Username      string
	Password      string
	Vhost         string
	Exchange      string
	KeyPrepared   string
	KeyCommited   string
	QueuePrepared string
	QueueCommited string
}

type RabbitmqChannel struct {
	Confirms chan amqp.Confirmation
	Channel  *amqp.Channel
}

type RabbitmqConst string

const (
	RabbitmqConstPrepared RabbitmqConst = "dtm_prepared"
	RabbitmqConstCommited RabbitmqConst = "dtm_commited"
)

func RabbitmqNew(conf *RabbitmqConfig) *Rabbitmq {
	return &Rabbitmq{
		Config: *conf,
		ChannelPool: &sync.Pool{
			New: func() interface{} {
				channel := newChannel(conf)
				err := channel.Confirm(false)
				common.PanicIfError(err)
				confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 2))
				return &RabbitmqChannel{
					Channel:  channel,
					Confirms: confirms,
				}
			},
		},
	}
}

func newChannel(conf *RabbitmqConfig) *amqp.Channel {
	uri := fmt.Sprintf("amqp://%s:%s@%s/%s", conf.Username, conf.Password, conf.Host, conf.Vhost)
	logrus.Printf("connecting rabbitmq: %s", uri)
	conn, err := amqp.Dial(uri)
	common.PanicIfError(err)
	channel, err := conn.Channel()
	common.PanicIfError(err)
	err = channel.ExchangeDeclare(
		conf.Exchange, // exchange name
		"direct",      // exchange type
		true,          // durable
		false,         // autoDelete
		false,         // internal
		false,         // noWait
		nil,           // args
	)
	common.PanicIfError(err)
	return channel
}

func (r *Rabbitmq) SendAndConfirm(key RabbitmqConst, data map[string]interface{}) error {
	body, err := json.Marshal(data)
	common.PanicIfError(err)
	channel := r.ChannelPool.Get().(*RabbitmqChannel)

	logrus.Printf("publishing %s %v", key, data)
	err = channel.Channel.Publish(
		r.Config.Exchange,
		common.If(key == RabbitmqConstPrepared, r.Config.KeyPrepared, r.Config.KeyCommited).(string),
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	common.PanicIfError(err)
	confirm := <-channel.Confirms
	r.ChannelPool.Put(channel)
	logrus.Printf("confirmed %t for %s", confirm.Ack, data["gid"])
	if !confirm.Ack {
		return fmt.Errorf("confirm not ok for %s", data["gid"])
	}
	return nil
}

type RabbitmqQueue struct {
	Name       string
	Queue      *amqp.Queue
	Channel    *amqp.Channel
	Conn       *amqp.Connection
	Deliveries <-chan amqp.Delivery
}

func (q *RabbitmqQueue) Close() {
	q.Channel.Close()
	// q.Conn.Close()
}

func (q *RabbitmqQueue) WaitAndHandle(handler func(data M)) {
	for {
		q.WaitAndHandleOne(handler)
	}
}
func (q *RabbitmqQueue) WaitAndHandleOne(handler func(data M)) {
	logrus.Printf("%s reading message", q.Name)
	msg := <-q.Deliveries
	data := map[string]interface{}{}
	err := json.Unmarshal(msg.Body, &data)
	logrus.Printf("%s handling one message: %v", q.Name, data)
	common.PanicIfError(err)
	handler(data)
	err = msg.Ack(false)
	common.PanicIfError(err)
	logrus.Printf("%s acked msg: %d", q.Name, msg.DeliveryTag)
}

func (r *Rabbitmq) QueueNew(queueType RabbitmqConst) *RabbitmqQueue {
	channel := newChannel(&r.Config)
	queueName := common.If(queueType == RabbitmqConstPrepared, r.Config.QueuePrepared, r.Config.QueueCommited).(string)
	queue, err := channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	common.PanicIfError(err)
	logrus.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange",
		queue.Name, queue.Messages, queue.Consumers)
	err = channel.QueueBind(
		queue.Name, // name of the queue
		common.If(queueType == RabbitmqConstPrepared, r.Config.KeyPrepared, r.Config.KeyCommited).(string), // bindingKey
		r.Config.Exchange, // sourceExchange
		false,             // noWait
		nil,               // arguments
	)
	common.PanicIfError(err)
	deliveries, err := channel.Consume(
		queue.Name,        // name
		"simple-consumer", // consumerTag,
		false,             // noAck
		false,             // exclusive
		false,             // noLocal
		false,             // noWait
		nil,               // arguments
	)
	common.PanicIfError(err)
	return &RabbitmqQueue{
		Queue:      &queue,
		Name:       queueName,
		Channel:    channel,
		Deliveries: deliveries,
	}
}

func (r *Rabbitmq) HandleMsg(data interface{}) {

}
