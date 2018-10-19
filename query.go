package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"time"
)

type AppealStreamMessage struct {
	AppealID string `json:"appealId"`
	ManagerId *string `json:"managerId"`
	Message string `json:"message"`
	Action string `json:"action"`
	ActionAt time.Time `json:"action_at"`
	ReceivedAt time.Time `json:"received_at"`
	Payload map[string]interface{}
}

func (currentClient *Client) query() {
	query, err := AMQPChannel.QueueDeclare(
		"erp_to_socket_message",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := AMQPChannel.Consume(
		query.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			appealStreamMessage := &AppealStreamMessage{}
			err := json.Unmarshal(d.Body, &appealStreamMessage)

			logger.WithFields(logrus.Fields{
				"error":  err,
			}).Error("Can`t decode query callBack:")

			bytesToSend, err := json.Marshal(appealStreamMessage)

			logger.WithFields(logrus.Fields{
				"error":  err,
			}).Error("Can`t encode query callBack:")

			currentClient.hub.broadcast <- bytesToSend

			logger.WithFields(logrus.Fields{
				"appeal":  appealStreamMessage.AppealID,
				"action":  appealStreamMessage.Action,
			}).Info("Read message from query:")

			d.Ack(false)
		}
	}()

	<-forever
}