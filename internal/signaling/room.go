package signaling

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

type broadcastMessage struct {
	sender *Client
	data   []byte
}

type messageBase struct {
	Type string `json:"type"`
}

type welcomeMessage struct {
	Type     string `json:"type"`
	ClientID string `json:"clientId"`
}

type participantsMessage struct {
	Type         string                  `json:"type"`
	Participants []participantDescriptor `json:"participants"`
}

type presenceMessage struct {
	Type     string `json:"type"`
	Action   string `json:"action"`
	ClientID string `json:"clientId"`
	TS       string `json:"ts"`
}

type profileMessage struct {
	Type        string `json:"type"`
	ClientID    string `json:"clientId"`
	DisplayName string `json:"displayName"`
}

type participantDescriptor struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName,omitempty"`
}

type Room struct {
	id         string
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMessage
	onEmpty    func(string)
	logger     *zap.Logger
}

func NewRoom(id string, onEmpty func(string), logger *zap.Logger) *Room {
	room := &Room{
		id:         id,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan broadcastMessage, 32),
		onEmpty:    onEmpty,
		logger:     logger,
	}

	go room.run()
	return room
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) Register(client *Client) {
	r.register <- client
}

func (r *Room) Unregister(client *Client) {
	r.unregister <- client
}

func (r *Room) Broadcast(sender *Client, data []byte) {
	r.broadcast <- broadcastMessage{sender: sender, data: data}
}

func (r *Room) HandleIncoming(sender *Client, data []byte) {
	var base messageBase
	if err := json.Unmarshal(data, &base); err == nil {
		if base.Type == "presence" || base.Type == "participants" || base.Type == "welcome" {
			return
		}

		if base.Type == "profile" {
			var msg profileMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				return
			}

			if msg.DisplayName == "" {
				return
			}

			r.broadcast <- broadcastMessage{
				sender: sender,
				data:   mustMarshal(profileMessage{Type: "profile", ClientID: sender.ID(), DisplayName: msg.DisplayName}),
			}

			return
		}
	}

	r.logger.Info("Received message", zap.String("roomID", r.ID()), zap.String("clientID", sender.ID()), zap.String("data", string(data)))
	r.Broadcast(sender, data)
}

func (r *Room) run() {
	clients := make(map[*Client]struct{})
	participants := make(map[string]struct{})
	displayNames := make(map[string]string)

	sendToClient := func(client *Client, payload interface{}) {
		data, err := json.Marshal(payload)
		if err != nil {
			return
		}

		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(clients, client)
		}
	}

	sendToAll := func(payload interface{}, skip *Client) {
		data, err := json.Marshal(payload)
		if err != nil {
			return
		}

		for client := range clients {
			if client == skip {
				continue
			}
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(clients, client)
			}
		}
	}

	for {
		select {
		case client := <-r.register:
			clients[client] = struct{}{}
			participants[client.ID()] = struct{}{}
			sendToClient(client, welcomeMessage{
				Type:     "welcome",
				ClientID: client.ID(),
			})
			ids := make([]participantDescriptor, 0, len(participants))
			for id := range participants {
				ids = append(ids, participantDescriptor{
					ID:          id,
					DisplayName: displayNames[id],
				})
			}
			sendToClient(client, participantsMessage{
				Type:         "participants",
				Participants: ids,
			})
			sendToAll(presenceMessage{
				Type:     "presence",
				Action:   "join",
				ClientID: client.ID(),
				TS:       time.Now().UTC().Format(time.RFC3339),
			}, client)
		case client := <-r.unregister:
			if _, ok := clients[client]; ok {
				delete(clients, client)
				close(client.send)
			}

			hasOther := false
			for c := range clients {
				if c.ID() == client.ID() {
					hasOther = true
					break
				}
			}

			if !hasOther {
				if _, ok := participants[client.ID()]; ok {
					delete(participants, client.ID())
					delete(displayNames, client.ID())
					sendToAll(presenceMessage{
						Type:     "presence",
						Action:   "leave",
						ClientID: client.ID(),
						TS:       time.Now().UTC().Format(time.RFC3339),
					}, nil)
				}
			}
			if len(clients) == 0 {
				if r.onEmpty != nil {
					r.onEmpty(r.id)
				}
				return
			}
		case msg := <-r.broadcast:
			if msg.sender != nil && len(msg.data) > 0 {
				var base messageBase
				if err := json.Unmarshal(msg.data, &base); err == nil && base.Type == "profile" {
					var profile profileMessage
					if err := json.Unmarshal(msg.data, &profile); err == nil {
						if profile.DisplayName != "" {
							displayNames[msg.sender.ID()] = profile.DisplayName
						}
					}
				}
			}
			for client := range clients {
				if client == msg.sender {
					continue
				}
				select {
				case client.send <- msg.data:
				default:
					close(client.send)
					delete(clients, client)
				}
			}
		}
	}
}

func mustMarshal(payload interface{}) []byte {
	data, err := json.Marshal(payload)
	if err != nil {
		return []byte("{}")
	}
	return data
}
