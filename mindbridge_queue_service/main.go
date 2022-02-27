package main

import (
	"fmt"
	"sync"

	configs "mindbridge_queue_service/config"
	Controller "mindbridge_queue_service/controller"
	"mindbridge_queue_service/esl"
	logging "mindbridge_queue_service/logger"
	queueRepo "mindbridge_queue_service/repository"
	queueUsecase "mindbridge_queue_service/usecases"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var onceRest sync.Once

func main() {
	fmt.Println("Welcome in queue service!")
	onceRest.Do(func() {
		e := echo.New()

		//Setting up the config
		config := configs.GetConfig()

		//Setting up the Logger
		logger := logging.NewLogger(config.Log.LogFile, config.Log.LogFile)

		e.Use(middleware.Logger())

		QueueRepo := queueRepo.NewQueueRepo()
		QueueUcase := queueUsecase.NewQueueUcase(QueueRepo)
		Controller.NewQueueController(e, QueueUcase)

		go esl.HandleEntryExitAnnounce()

		fmt.Println(config.HttpConfig.HostPort)
		if err := e.Start(config.HttpConfig.HostPort); err != nil {
			fmt.Println(err)
			logger.WithError(err).Fatal("Unable to start the callCenter service without tls")
		}

	})
}
