package esl

import (
	"encoding/json"
	"flag"
	"fmt"
	"mindbridge_queue_service/models"

	. "github.com/0x19/goesl"
	"github.com/streadway/amqp"
)

var (
	fshost   = flag.String("fshost", "localhost", "Freeswitch hostname. Default: localhost")
	fsport   = flag.Uint("fsport", 8021, "Freeswitch port. Default: 8021")
	password = flag.String("pass", "ClueCon", "Freeswitch password. Default: ClueCon")
	timeout  = flag.Int("timeout", 10, "Freeswitch conneciton timeout in seconds. Default: 10")
)

func HandleEntryExitAnnounce() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Failed Initializing Broker Connection")
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	if err != nil {
		fmt.Println(err)
	}

	msgs, err := ch.Consume(
		"Queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			var obj models.Entry

			if err := json.Unmarshal(d.Body, &obj); err != nil {
				//if err := json.Unmarshal(d.Body, &entry); err != nil {
				panic(err)
				//}
			}
			//fmt.Println("Recieved Message:", d.Body)
			fmt.Println("In handleEntryExitAnnounce New msg comes in   ", obj)

			if obj.EntryAnnounce == true {
				client, err := NewClient(*fshost, *fsport, *password, *timeout)

				if err != nil {
					Error("Error while creating new client: %s", err)
					return
				}
				client.Send("events json ALL")

				client.BgApi(fmt.Sprintf("conference %s %s", obj.ConferenceName, "moh off"))
				client.BgApi(fmt.Sprintf("conference %s %s", obj.ConferenceName, "play /usr/local/freeswitch/prompt/04.mp3"))
				client.BgApi(fmt.Sprintf("conference %s %s", obj.ConferenceName, "play /tmp/"+obj.ConferenceUUID+".mp3"))
				client.BgApi(fmt.Sprintf("conference %s %s", obj.ConferenceName, "play /usr/local/freeswitch/prompt/beep.wav"))

				fmt.Println("After playing entry announce")

				//if client.Close().Error() != "" {
				//	fmt.Println("Pronblem to in close the connection")
				//}
				client.Exit()
			} else if obj.ConferenceName != "" {
				client, err := NewClient(*fshost, *fsport, *password, *timeout)

				if err != nil {
					Error("Error while creating new client: %s", err)
					return
				}
				client.Send("events json ALL")

				client.BgApi(fmt.Sprintf("conference %s %s", obj.ConferenceName, "play /usr/local/freeswitch/prompt/now-exiting.wav"))
				client.BgApi(fmt.Sprintf("conference %s %s", obj.ConferenceName, "play /tmp/"+obj.ConferenceUUID+".mp3"))
				client.BgApi(fmt.Sprintf("conference %s %s", obj.ConferenceName, "play /usr/local/freeswitch/prompt/beep.wav"))
				client.BgApi(fmt.Sprintf("conference %s %s", obj.ConferenceName, "moh off"))
				fmt.Println("After playing exit announce")

				//if client.Close().Error() != "" {
				//	fmt.Println("Pronblem to in close the connection")
				//}
				client.Exit()
			} else {
				fmt.Println("Nothing to play")
			}
			//client.BgApi(fmt.Sprintf("originate %s %s", "sofia/internal/1001@192.168.43.61", "&playback(/usr/local/freeswitch/prompt/hold_music.wav)"))
		}
	}()

	fmt.Println("Successfully Connected to our RabbitMQ Instance")
	fmt.Println(" [*] - Waiting for messages")
	<-forever

	//client.BgApi(fmt.Sprintf("originate %s %s", "sofia/internal/1001@192.168.43.61", "&playback(/usr/local/freeswitch/prompt/hold_music.wav)"))

	fmt.Println("Outbound call placed")

}
