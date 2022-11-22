package websocketrooms

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents the websocket client at the server
type Client struct {
	// The actual websocket connection.
	conn     *websocket.Conn
	wsServer *WsServer
	send     chan []byte
	ID       string `json:"id"`
	rooms    map[*Room]bool
}

func newClient(conn *websocket.Conn, wsServer *WsServer, userID string) *Client {
	return &Client{
		ID:       userID,
		conn:     conn,
		wsServer: wsServer,
		send:     make(chan []byte, 256),
		rooms:    make(map[*Room]bool),
	}

}

func (client *Client) readPump() {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		client.handleNewMessage(jsonMessage)
	}

}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	client.wsServer.unregister <- client
	for room := range client.rooms {
		room.unregister <- client
	}
	close(client.send)
	client.conn.Close()
}

// ServeWs handles websocket requests from clients requests.
func ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request, userID string) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, wsServer, userID)

	go client.writePump()
	go client.readPump()

	wsServer.register <- client
}

func (client *Client) handleNewMessage(jsonMessage []byte) {

	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
		return
	}

	message.Sender = client
	message.Time = time.Now()

	switch message.Action {
	case SendMessageAction:
		roomID := message.Target
		if room := client.wsServer.findRoomByID(roomID); room != nil {
			room.broadcast <- &message
		}

		// var messageDB = models.Message{ID: uuid.New().String(), RoomID: roomID, Sender: message.Sender.ID, Message: string(message.Message), CreatedAt: time.Now()}
		// initializers.DB.Save(&messageDB)
		models.SendMessage(initializers.DB, uuid.New().String(), roomID, message.Sender.ID, string(message.Message), time.Now())

		chatUser1, chatUser2 := models.FindChatRoomUsers(initializers.DB, roomID)
		// initializers.DB.Table("rooms").Select("user1_id").Where("id = ?", roomID).Find(&chatUser1)
		// initializers.DB.Table("rooms").Select("user2_id").Where("id = ?", roomID).Find(&chatUser2)
		if message.Sender.ID == chatUser1 {
			models.SaveNotification(initializers.DB, chatUser2, message.Sender.ID, "sent you a message", "/chat")
		} else {
			models.SaveNotification(initializers.DB, chatUser1, message.Sender.ID, "sent you a message", "/chat")
		}
	case LeaveRoomAction:
		client.handleLeaveRoomMessage(message)

	case JoinRoomPrivateAction:
		client.handleJoinRoomPrivateMessage(message)
	}
}

func (client *Client) handleLeaveRoomMessage(message Message) {
	room := client.wsServer.findRoomByID(message.Target)
	if room == nil {
		return
	}

	delete(client.rooms, room)
	room.unregister <- client
}

func (client *Client) handleJoinRoomPrivateMessage(message Message) {

	target := client.wsServer.findClientByID(message.Sender.ID)

	if target == nil {
		return
	}

	// create unique room name combined to the two IDs
	roomID := message.Target

	client.joinRoom(roomID, target)
	target.joinRoom(roomID, client)
}

func (client *Client) joinRoom(roomID string, sender *Client) {

	room := client.wsServer.findRoomByID(roomID)
	if room == nil {
		room = client.wsServer.createRoom(roomID)
	}

	if !client.isInRoom(room) {

		client.rooms[room] = true
		room.register <- client
	}

}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room]; ok {
		return true
	}
	return false
}

func (client *Client) GetID() string {
	return client.ID
}

// serveWsForNotifications handles websocket for notifications
func ServeWsForNotifications(wsServer *WsServer, w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, wsServer, userID)
	wsServer.register <- client

	for {
		number := models.GetNumberOfNotifications(initializers.DB, userID)

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write([]byte(fmt.Sprint(number)))

		if err := w.Close(); err != nil {
			return
		}
		time.Sleep(5 * time.Second)

	}

}
