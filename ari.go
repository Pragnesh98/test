package main

import (
	"context"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"

	"bitbucket.org/yellowmessenger/asterisk-ari/bothelper"
	"bitbucket.org/yellowmessenger/asterisk-ari/callback"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/connections"
	"bitbucket.org/yellowmessenger/asterisk-ari/enqueuecallworker"
	"bitbucket.org/yellowmessenger/asterisk-ari/eventhandler"
	"bitbucket.org/yellowmessenger/asterisk-ari/globals"

	"bitbucket.org/yellowmessenger/asterisk-ari/metrics"
	"bitbucket.org/yellowmessenger/asterisk-ari/models/mysql"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/queuemanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/spanhealth"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/azure"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/yellowmessenger"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/speechtotext"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v3"
	echopprof "github.com/sevenNt/echo-pprof"
)

var (
	host = "0.0.0.0"
	port = "9991"
)

func main() {
	// Initialize new relic app
	if err := newrelic.InitNewRelicApp(); err != nil {
		log.Fatalf("Error while initializing new relic app. Error: [%#v]", err)
		panic(1)
	}
	e := echo.New()
	// Set the middlewares
	// Register new relic middleware
	e.Use(nrecho.Middleware(newrelic.App))
	e.Use(middleware.Secure())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("1024KB"))
	e.Use(middleware.RemoveTrailingSlash())
	// Set the logging
	loggerConfig := middleware.DefaultLoggerConfig
	// Uncomment to write the logs to a file
	// file, err := os.OpenFile("/var/log/yellowmessenger/asterisk_ari/access.log", os.O_APPEND|os.O_WRONLY, 0644)
	// if err != nil {
	// 	log.Fatalf("Failed to open the file. Error: [%#v]", err)
	// }
	// loggerConfig.Output = file
	e.Debug = true
	e.Use(middleware.LoggerWithConfig(loggerConfig))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Initilize the config
	if err := configmanager.InitConfig("config.json"); err != nil {
		log.Fatalf("Error while initializing the config. Error: [%#v]", err)
		panic(1)
	}

	// Initiliaze YM logger
	if err := ymlogger.InitYMLogger(configmanager.ConfStore.LoggerConf); err != nil {
		log.Fatalf("Failed to initialize the logger. Err: [%#v]", err)
	}

	// Generate Google Token periodically
	go configmanager.RenewGoogleToken(ctx)

	// Generate Azure TTS Token Periodically
	go azure.RenewAzureTTSToken(ctx)

	// Initialize MySQL Connection
	if err := mysql.Init(); err != nil {
		log.Fatalf("Failed to initialize MySQL Connection. Error: [%#v]", err)
	}
	// Initialize Metrics client
	if err := metrics.InitClient(configmanager.ConfStore.MetricsConf); err != nil {
		log.Fatalf("Failed to initialize metrics client")
	}

	// Initialize BOT HTTP Client
	bothelper.InitBotHTTPclient()

	// Initialize Azure STT HTTP Client
	azure.InitAzureSTTHTTPClient()

	// Initialize yellowmessenger STT HTTP Client
	yellowmessenger.InitYMSTTHTTPClient()

	// Initialize Azure TTS HTTP Client
	azure.InitAzureTTSHTTPClient()

	// Initialize Call Store HTTP Client
	callstore.InitCallStoreClient()

	// Start PipeHealth
	go spanhealth.InitSpanHealth()

	// Connect to ARI
	ariClient, err := connections.ConnectARI(ctx)
	if err != nil {
		ymlogger.LogErrorf("ARIConnect", "Error while connecting to ARI. Error: [%#v]", err)
	}
	// Initialize the handler
	ymlogger.LogInfo("InitHandler", "Going to start the channel handler")
	go eventhandler.InitHandler(ctx, ariClient, eventhandler.ChannelHandler)
	// Initialize Channel Operation handlers
	ymlogger.LogInfo("InitHandler", "Going to start the channel operation handler")
	go eventhandler.InitChannelOpHandler(ctx, ariClient)

	// Initialize RabbitMQ Connection
	ymlogger.LogInfo("InitRabbitMQConn", "Initializing RabbitMQ Connection")
	if err := queuemanager.InitRabbitMQConn(configmanager.ConfStore.QueueConnParams); err != nil {
		log.Fatalf("Failed to initialize Rabbit MQ Connection. Error: [%#v]", err)
	}

	// Start Queueworker
	ymlogger.LogInfo("InitRabbitMQueueListneter", "Initializing RabbitMQ Queue Listener")
	if err := queuemanager.InitQueueListener(
		configmanager.ConfStore.QueueListenerParams,
		&enqueuecallworker.EnqueueCallWorker{},
		configmanager.ConfStore.CampaignDelayPerCallMS,
		configmanager.ConfStore.CampaignMinHour,
		configmanager.ConfStore.CampaignMaxHour,
		&configmanager.ConfStore.BotRateLimitParams); err != nil {
		log.Fatalf("Failed to initialize queue listener. Error: [%#v]", err)
	}

	// Initialize BOT HTTP Client
	bothelper.InitBotHTTPclient()

	// Initialize Azure HTTP Client
	azure.InitAzureSTTHTTPClient()

	// Initialize Callback HTTP Client
	callback.InitCallbackClient()

	// Intialize Call counter
	globals.InitCounter()

	// Start callbacks
	go callback.StartWorker(ctx)

	// Initialize GRPC connection
	if err := speechtotext.InitGRPCConnNew(); err != nil {
		log.Fatalf("Failed to initialize GRPC connection. Error: [%#v]", err)
	}

	// AddingRoutes
	AddRoutes(e)

	// Add pprof
	echopprof.Wrap(e)

	ymlogger.LogInfof("HTTPHandler", "Listening for requests on port %s", port)
	if err := e.Start(fmt.Sprintf("%s:%s", host, port)); err != nil {
		ymlogger.LogCritical("HTTPHandler", "Failed to start server!", err)
		os.Exit(1)
	}
}
