package queuemanager

import (
	"fmt"
	"strconv"

	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/streadway/amqp"
)

type QueueConnParams struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	UserName       string `json:"user_name"`
	Password       string `json:"password"`
	QueueName      string `json:"queue_name"`
	Durable        bool   `json:"durable"`
	DeleteUnused   bool   `json:"delete_unused"`
	Exclusive      bool   `json:"exclusive"`
	NoWait         bool   `json:"no_wait"`
	TTL            int    `json:"ttl"`
	MaxQueueLength int    `json:"max_queue_length"`
}

type QueueMessageParams struct {
	Exchange  string `json:"exchange"`
	QueueName string `json:"queue_name"`
	Msg       string `json:"msg"`
	Delay     int64  `json:"delay"`
	Priority  uint8  `json:"priority"`
	Mandatory bool   `json:"mandatory"`
	Immediate bool   `json:"immediate"`
}

var ch *amqp.Channel

func InitRabbitMQConn(qParams QueueConnParams) error {
	conn, err := amqp.Dial("amqp://" + fmt.Sprintf("%s:%s@%s:%s", qParams.UserName, qParams.Password, qParams.Host, strconv.Itoa(qParams.Port)))
	if err != nil {
		ymlogger.LogErrorf("InitRabbitMQ", "Failed to connect to RabbitMQ. Error: [%#v]", err)
		return err
	}

	ch, err = conn.Channel()
	if err != nil {
		ymlogger.LogErrorf("InitRabbitMQ", "Failed to open a channel. Error: [%#v]", err)
		return err
	}
	// declare exchange if not exist
	eargs := make(amqp.Table)
	eargs["x-delayed-type"] = "direct"
	err = ch.ExchangeDeclare("delayed", "x-delayed-message", true, false, false, false, eargs)
	if err != nil {
		ymlogger.LogErrorf("InitRabbitMQ", "Failed to declare the exchange. Error: [%#v]", err)
		return err
	}
	var args = make(amqp.Table)
	args["x-max-priority"] = 10
	if qParams.TTL > 0 {
		args["x-message-ttl"] = qParams.TTL
	}
	if qParams.MaxQueueLength > 0 {
		args["x-max-length"] = qParams.MaxQueueLength
	}
	q, err := ch.QueueDeclare(
		qParams.QueueName,
		qParams.Durable,
		qParams.DeleteUnused,
		qParams.Exclusive,
		qParams.NoWait,
		args,
	)
	if err != nil {
		ymlogger.LogErrorf("InitRabbitMQ", "Failed to declare the queue. Error: [%#v]", err)
		return err
	}
	err = ch.QueueBind(q.Name, "enqueuecall", "delayed", false, nil)
	if err != nil {
		ymlogger.LogErrorf("InitRabbitMQ", "Failed to bind the queue. Error: [%#v]", err)
		return err
	}
	err = ch.Qos(2, 0, false)
	if err != nil {
		return err
	}
	ymlogger.LogDebugf("QueueStats", "QueueName: [%s] NumOfMessages: [%d]", q.Name, q.Messages)
	return nil
}

func (param *QueueMessageParams) Enqueue() error {
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(param.Msg),
	}
	if param.Priority > 0 {
		msg.Priority = param.Priority
	}
	if param.Delay > 0 {
		headers := make(amqp.Table)
		headers["x-delay"] = param.Delay
		headers["x-delayed-type"] = "direct"
		headers["x-delayed-message"] = true
		msg.Headers = headers
	}
	err := ch.Publish(
		param.Exchange,
		param.QueueName,
		param.Mandatory,
		param.Immediate,
		msg,
	)
	if err != nil {
		ymlogger.LogErrorf("EnqueueMsg", "Error while enqueuing the msg. Msg: [%s] Error: [%#v]", param.Msg, err)
		return err
	}
	return nil
}
