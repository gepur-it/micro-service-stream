package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type AppealStreamMessage struct {
	AppealID   string  `json:"appealId"`
	ManagerId  *string `json:"managerId"`
	Message    string  `json:"message"`
	Action     string  `json:"action"`
	ActionAt   string  `json:"actionAt"`
	ReceivedAt string  `json:"receivedAt"`
}

func (h *Hub) query() {
	msgs, err := AMQPChannel.Consume(
		"erp_to_socket_appeal",
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

			if err != nil {
				logger.WithFields(logrus.Fields{
					"error": err,
				}).Error("Can`t decode query callBack:")
			}

			bytesToSend, err := json.Marshal(appealStreamMessage)

			if err != nil {
				logger.WithFields(logrus.Fields{
					"error": err,
				}).Error("Can`t encode query callBack:")
			}

			h.broadcast <- bytesToSend

			logger.WithFields(logrus.Fields{
				"appeal": appealStreamMessage.AppealID,
				"action": appealStreamMessage.Action,
			}).Info("Read message from query:")

			d.Ack(false)
		}
	}()

	<-forever
}
