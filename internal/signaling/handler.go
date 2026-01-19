package signaling

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"go.uber.org/zap"
)

type TokenParser interface {
	ParseAccessToken(tokenString string) (string, uint64, error)
}

type Handler struct {
	registry    *Registry
	logger      *zap.Logger
	tokenParser TokenParser
}

func NewHandler(registry *Registry, tokenParser TokenParser, logger *zap.Logger) *Handler {
	return &Handler{
		registry:    registry,
		logger:      logger,
		tokenParser: tokenParser,
	}
}

type createRoomResponse struct {
	RoomID string `json:"roomId"`
	WsURL  string `json:"wsUrl"`
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := generateRoomID()
	if err != nil {
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	resp := createRoomResponse{
		RoomID: roomID,
		WsURL:  websocketURL(r, roomID),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/ws/")
	if roomID == "" || strings.Contains(roomID, "/") {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}

	var (
		clientName   string
		clientUserID uint64
	)

	token := r.URL.Query().Get("token")
	if token != "" {
		if username, userID, err := h.tokenParser.ParseAccessToken(token); err == nil && username != "" && userID != 0 {
			clientName = username
			clientUserID = userID
		}
	}

	if clientName == "" {
		clientName = randomID()
	}

	room := h.registry.GetOrCreate(roomID)
	client := NewClient(clientUserID, clientName, conn, room)
	room.Register(client)

	client.Run(r.Context())

	h.logger.Info("Client joined room", zap.String("roomID", roomID), zap.Uint64("clientUserID", client.ID()))
}

func randomID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}

	return hex.EncodeToString(buf)
}

func generateRoomID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func websocketURL(r *http.Request, roomID string) string {
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}

	host := r.Host
	return scheme + "://" + host + "/ws/" + roomID
}
