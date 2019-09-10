package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type QueryCommand struct {
	Module     string                 `json:"module"`
	Action     string                 `json:"action"`
	Type       string                 `json:"type"`
	Recipients []string               `json:"recipients"`
	Payload    map[string]interface{} `json:"payload"`
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
    notify := AMQPConnection.NotifyClose(make(chan *amqp.Error))

	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
	    for {
	        select {
	            case err = <- notify:
	                failOnError(err, "Queue is dead")
	                break
	            case d := <-msgs:
	                h.mu.Lock()
              		logger.Info("Query receive message:")
              		queryCommand := QueryCommand{}

              		err = json.Unmarshal(d.Body, &queryCommand)

              		if err != nil {
              			logger.WithFields(logrus.Fields{
                  			"error": err,
                		}).Error("Can`t decode query command:")
                	}

                	if queryCommand.Type == "subscribe" {
                		h.subscribe <- queryCommand
                		d.Ack(false)
                	} else {
                		h.broadcast <- d.Body
                		d.Ack(false)
                	}

        			h.mu.Unlock()
	        }
	    }
	}()

	<-forever
}
