package main

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
			logger.Info("Query receive message:")

			h.broadcast <- d.Body
			d.Ack(false)
		}
	}()

	<-forever
}
