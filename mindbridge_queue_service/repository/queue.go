package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mindbridge_queue_service/interfaces"
	"mindbridge_queue_service/models"
	"net/http"
	"os/exec"

	"github.com/streadway/amqp"
)

/*type body struct {
	conferenceUUID string `json:"conference_uuid"`
	conferenceName string `json:"conference_name"`
}*/

type QueueRepo struct{}

func NewQueueRepo() interfaces.QueueRepoInterface {
	return &QueueRepo{}
}

func (r *QueueRepo) HandleQueue(ctx context.Context, conferenceUUID, conferenceName string) (*models.QueueApiResponse, error) {

	conference := models.Body{
		ConferenceUUID: conferenceUUID,
		ConferenceName: conferenceName,
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(conference)

	fmt.Println("In repository delete", conferenceName, conferenceUUID)
	if conferenceName != "nil" {

		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(conference)

		//reqBodyBytes.Bytes()
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			fmt.Println(err)
			panic(1)
		}
		defer conn.Close()

		fmt.Println("Going to publish msg in RabbitMQ Instance for exit announce")
		filename := conferenceUUID + ".mp3"
		cmd, err := exec.Command("/bin/bash", "/home/ubuntu/volume.sh", filename).Output()
		s := string(cmd)
		fmt.Println("cmd=================>", s)

		if err != nil {
			fmt.Println(err)
		}

		ch, err := conn.Channel()
		if err != nil {
			fmt.Println(err)
		}
		defer ch.Close()

		q, err := ch.QueueDeclare(
			"Queue",
			false,
			false,
			false,
			false,
			nil,
		)

		fmt.Println(q, reqBodyBytes.Bytes())

		if err != nil {
			fmt.Println(err)
		}

		err = ch.Publish(
			"",
			"Queue",
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(reqBodyBytes.Bytes()),
			},
		)
		if err != nil {
			fmt.Println(err)
		}
	}

	return &models.QueueApiResponse{Status: "1", Msg: "Successfully Published Message to Queue for exit announce", ResponseCode: http.StatusOK}, nil

}

func (r *QueueRepo) HandleEntryQueue(ctx context.Context, conferenceUUID, conferenceName string) (*models.QueueApiResponse, error) {

	entry := true
	conference := models.Entry{
		EntryAnnounce:  entry,
		ConferenceUUID: conferenceUUID,
		ConferenceName: conferenceName,
	}

	fmt.Println("In Entry announce", conferenceName, conferenceUUID, entry)
	if conferenceName != "nil" {

		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(conference)

		//reqBodyBytes.Bytes()
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			fmt.Println(err)
			panic(1)
		}
		defer conn.Close()

		fmt.Println("Going to publish body in RabbitMQ Instance for entry announce")

		filename := conferenceUUID + ".mp3"
		cmd, err := exec.Command("/bin/bash", "/home/ubuntu/volume.sh", filename).Output()
		s := string(cmd)
		fmt.Println("cmd=================>", s)

		if err != nil {
			fmt.Println(err)
		}

		ch, err := conn.Channel()
		if err != nil {
			fmt.Println(err)
		}
		defer ch.Close()

		q, err := ch.QueueDeclare(
			"Queue",
			false,
			false,
			false,
			false,
			nil,
		)

		fmt.Println(q, reqBodyBytes.Bytes())

		if err != nil {
			fmt.Println(err)
		}

		err = ch.Publish(
			"",
			"Queue",
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(reqBodyBytes.Bytes()),
			},
		)
		if err != nil {
			fmt.Println(err)
		}
	}

	return &models.QueueApiResponse{Status: "1", Msg: "Successfully Published Message to Queue for entry", ResponseCode: http.StatusOK}, nil

}

func (r *QueueRepo) DeleteQueue(ctx context.Context) (*models.QueueApiResponse, error) {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err, "Failed to open a channel")
	}
	defer ch.Close()

	_, declare_err := ch.QueueDeclare(
		"Queue", // name
		true,    // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		fmt.Println(declare_err, "Failed to declare a queue")
	}

	purged, err := ch.QueuePurge("Queue", false)
	if err != nil {
		fmt.Println(err, "Failed to register a consumer")
	}
	fmt.Println("%d messages purged from queue", purged)

	return &models.QueueApiResponse{Status: "1", Msg: "Delete queue succesfully", ResponseCode: http.StatusOK}, nil
}
