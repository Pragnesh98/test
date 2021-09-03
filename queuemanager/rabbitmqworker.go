package queuemanager

import (
	"context"
	"sync"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/utils/ratelimit"

	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/streadway/amqp"
)

type JobStatus string

const (
	Success     JobStatus = "success"
	Failure               = "failure"
	TempFailure           = "temporary_failure"
)

type IQWorker interface {
	Process(jobMsg []byte, botRateLimit *BotRateLimits) QueueJobResult
}

type QueueListenerParams struct {
	QueueName  string `json:"queue_name"`
	AutoAck    bool   `json:"auto_ack"`
	Exclusive  bool   `json:"exclusive"`
	NoLocal    bool   `json:"no_local"`
	NoWait     bool   `json:"no_wait"`
	NumWorkers int    `json:"num_workers"`
}
type QueueJobResult struct {
	Status   JobStatus
	Delay    int64
	Priority uint8
}

type BotRateLimitConfig struct {
	PhoneNumber      string  `json:"phone_number"`
	RateLimitEnabled bool    `json:"rate_limit_enabled"`
	ThresholdMillis  int     `json:"threshold_millis"`
	MinHour          int     `json:"min_hour"`
	MaxHour          int     `json:"max_hour"`
	MaxRate          float64 `json:"max_rate"`
	Burst            int     `json:"burst"`
}

type BotRateLimitParams struct {
	RateLimitConfigs map[string]BotRateLimitConfig `json:"rate_limit_configs"`
}

type BotRateLimits struct {
	mu           sync.Mutex
	ratelimitMap map[string]*ratelimit.AdaptiveRateLimiter
	conf         BotRateLimitParams
}

func (b *BotRateLimits) ensureRateLimiter(phoneNumber string) *ratelimit.AdaptiveRateLimiter {
	b.mu.Lock()
	defer b.mu.Unlock()

	var threshold = time.Second * 3
	if botConf, ok := b.conf.RateLimitConfigs[phoneNumber]; ok {
		threshold = time.Millisecond * time.Duration(botConf.ThresholdMillis)
	}
	if _, ok := b.ratelimitMap[phoneNumber]; !ok {
		var maxRate float64 = 2
		var burst int = 2
		if botConf, ok := b.conf.RateLimitConfigs[phoneNumber]; ok {
			if botConf.MaxRate != 0 {
				maxRate = botConf.MaxRate
			}
			if botConf.Burst != 0 {
				burst = botConf.Burst
			}
		}
		b.ratelimitMap[phoneNumber] = ratelimit.New(maxRate, burst, threshold, phoneNumber)
	}

	return b.ratelimitMap[phoneNumber]
}

func (b *BotRateLimits) Wait(ctx context.Context, phoneNumber string) {
	if b == nil {
		ymlogger.LogErrorf("BotRateLimits", "Wait called with nil")
		return
	}
	startTime := time.Now()

	ymlogger.LogDebugf("BotRateLimits", "%s: waiting for ratelimit", phoneNumber)
	defer func() {
		ymlogger.LogDebugf("BotRateLimits", "%s: waited for %d millis", phoneNumber,
			time.Since(startTime).Milliseconds())
	}()

	rateLimit := b.ensureRateLimiter(phoneNumber)
	if botConf, ok := b.conf.RateLimitConfigs[phoneNumber]; ok {
		if !botConf.RateLimitEnabled {
			ymlogger.LogDebugf("BotRateLimits", "Rate Limiting is disabled for [%s]", phoneNumber)
			return
		}
	}

	rateLimit.Wait(ctx)
}

func (b *BotRateLimits) GetBotRateLimiter(phoneNumber string) *ratelimit.AdaptiveRateLimiter {
	return b.ensureRateLimiter(phoneNumber)
}

func (b *BotRateLimits) GetBotRateLimitConf(phoneNumber string) *BotRateLimitConfig {
	if botConf, ok := b.conf.RateLimitConfigs[phoneNumber]; ok {
		return &botConf
	}
	return nil
}

func InitQueueListener(params QueueListenerParams, worker IQWorker, campaignDelay int, minHour int, maxHour int, rateLimitParams *BotRateLimitParams) error {
	messages, err := ch.Consume(
		params.QueueName,
		"enqueuecallconsumer",
		params.AutoAck,
		params.Exclusive,
		params.NoLocal,
		params.NoWait,
		nil,
	)
	if err != nil {
		ymlogger.LogErrorf("QueueListener", "Failed to start consuming the messages. Error: [%#v]", err)
		return err
	}
	StartWorkers(worker, params.NumWorkers, messages, campaignDelay, minHour, maxHour, rateLimitParams)
	return nil
}

func StartWorkers(worker IQWorker, numWorkers int, messages <-chan amqp.Delivery, campaignDelay int, minHour int, maxHour int, rateLimitParams *BotRateLimitParams) {
	botRateLimits := &BotRateLimits{}
	botRateLimits.ratelimitMap = make(map[string]*ratelimit.AdaptiveRateLimiter)
	botRateLimits.conf = *rateLimitParams
	for i := 1; i <= numWorkers; i++ {
		go processJob(i, worker, messages, campaignDelay, minHour, maxHour, botRateLimits)
	}
}

func processJob(numWorker int, worker IQWorker, messages <-chan amqp.Delivery, campaignDelay int, minHour int, maxHour int, botRateLmits *BotRateLimits) {
	for message := range messages {
		// Sleep for the delay specified
		time.Sleep(time.Duration(campaignDelay) * time.Millisecond)
		//check on theshold value. How many concurrent calls running on server
		loc, _ := time.LoadLocation("Asia/Kolkata")
		if time.Now().In(loc).Hour() > maxHour || time.Now().In(loc).Hour() < minHour {
			ymlogger.LogInfo("QueueListener", "No campaign allowed during this time. Message Body: [%#s]. Re-enqueueing..", string(message.Body))
			result := QueueJobResult{
				Status:   TempFailure,
				Priority: 9,
				Delay:    100000, // in MS
			}
			finishTask(result, message)
			continue
		}
		go func(message amqp.Delivery) {
			result := worker.Process(message.Body, botRateLmits)
			finishTask(result, message)
		}(message)
	}
}

func finishTask(
	jobResult QueueJobResult,
	message amqp.Delivery,
) {
	
	if err := message.Ack(false); err != nil {
		ymlogger.LogErrorf("QueueListener", "Error acknowledging message : %s", err)
	} else {
		ymlogger.LogInfof("QueueListener", "Acknowledged message [%#v]", string(message.Body))
	}

	switch jobResult.Status {
	case Success:
		fallthrough
	case Failure:
		return
	}
	job, err := formEnqueueMessage(jobResult, message)
	if err != nil {
		ymlogger.LogErrorf("QueueListener", "Failed to form enqueue message. Message: [%s] Error: [%#v]", string(message.Body), err)
		return
	}
	job.Enqueue()
}

func formEnqueueMessage(
	jobResult QueueJobResult,
	message amqp.Delivery,
) (QueueMessageParams, error) {
	job := QueueMessageParams{
		Exchange:  "delayed",
		QueueName: message.RoutingKey,
		Msg:       string(message.Body),
		Mandatory: true,
		Immediate: false,
	}
	if jobResult.Delay > 0 {
		job.Delay = jobResult.Delay
	}
	if jobResult.Priority > 0 {
		job.Priority = jobResult.Priority
	}
	return job, nil
}
