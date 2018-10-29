package main

import (
	"bytes"
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"time"
)

type User struct {
	Id           bson.ObjectId `json:"id"        bson:"_id,omitempty"`
	Username     string        `json:"username" bson:"username"`
	UserID       string        `json:"userId" bson:"userId"`
	ObjectGUID   string        `json:"objectGUID" bson:"objectGUID"`
	LastActivity time.Time     `json:"lastActivity" bson:"lastActivity"`
}

type SocketResponse struct {
	StatusMessage string `json:"statusMessage"`
	StatusCode    int    `json:"statusCode"`
}

func (currentClient *Client) readPump() {
	defer func() {
		currentClient.hub.unregister <- currentClient
		currentClient.conn.Close()
	}()

	currentClient.conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := currentClient.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.WithFields(logrus.Fields{
					"error": err,
					"addr":  currentClient.conn.RemoteAddr(),
				}).Error("Socket error:")
			}

			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		if string(message) != "." {
			logger.WithFields(logrus.Fields{
				"addr": currentClient.conn.RemoteAddr(),
			}).Info("Socket receive message:")

			if len(currentClient.subscribe.Id) == 0 {
				subscribe := Subscribe{}
				err = json.Unmarshal(message, &subscribe)

				if err != nil || len(subscribe.ApiKey) == 0 {

					logger.WithFields(logrus.Fields{
						"addr": currentClient.conn.RemoteAddr(),
					}).Warn("Hasta la vista, baby! Can`t decode subscribe message:")

					currentClient.hub.unregister <- currentClient

					logger.WithFields(logrus.Fields{
						"addr": currentClient.conn.RemoteAddr(),
					}).Warn("Unregister current client:")

					break
				}

				user := User{}

				err = MgoCollection.Find(bson.M{"_id": subscribe.ApiKey}).One(&user)

				if err != nil {
					logger.WithFields(logrus.Fields{
						"error": err,
						"addr":  currentClient.conn.RemoteAddr(),
					}).Warn("Hasta la vista, baby! User %s not accept to connect this chat:")

					break
				}

				socketResponse := SocketResponse{StatusMessage: "ok", StatusCode: 200}

				bytesToSend, err := json.Marshal(socketResponse)

				if err != nil {
					logger.WithFields(logrus.Fields{
						"error": err,
					}).Error("Can`t encode socket response:")
				}

				err = setStatus(user.UserID, true)

				if err != nil {
					logger.WithFields(logrus.Fields{
						"manager": user.UserID,
						"err":     err,
					}).Error("Manager can`t update status:")

					return
				}

				currentClient.subscribe = user

				logger.WithFields(logrus.Fields{
					"user": user.Username,
				}).Info("User registered:")
				currentClient.send <- bytesToSend

				logger.WithFields(logrus.Fields{
					"addr": currentClient.conn.RemoteAddr(),
				}).Info("Client subscribe:")
			}
		} else {
			err = setStatus(currentClient.subscribe.UserID, true)

			if err != nil {
				logger.WithFields(logrus.Fields{
					"manager": currentClient.subscribe.UserID,
					"err":     err,
				}).Error("Manager can`t update status:")

				return
			}

			logger.WithFields(logrus.Fields{
				"addr": currentClient.conn.RemoteAddr(),
			}).Info("Recv pong:")
		}
	}
}

func (currentClient *Client) writePump() {

	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		currentClient.conn.Close()
	}()

	for {
		select {
		case message, ok := <-currentClient.send:
			currentClient.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// The hub closed the channel.
				currentClient.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := currentClient.conn.NextWriter(websocket.TextMessage)

			if err != nil {
				return
			}

			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(currentClient.send)

			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-currentClient.send)
			}

			if err := w.Close(); err != nil {
				return
			}

			logger.WithFields(logrus.Fields{
				"addr": currentClient.conn.RemoteAddr(),
			}).Info("Socket send message to client:")

		case <-ticker.C:
			currentClient.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := currentClient.conn.WriteMessage(websocket.TextMessage, []byte(".")); err != nil {
				logger.WithFields(logrus.Fields{
					"error": err,
					"addr":  currentClient.conn.RemoteAddr(),
				}).Error("Ping send error:")

				return
			}

			logger.WithFields(logrus.Fields{
				"addr": currentClient.conn.RemoteAddr(),
			}).Info("Ping send:")
		}
	}
}
