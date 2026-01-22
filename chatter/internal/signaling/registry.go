package signaling

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type Registry struct {
	mu     sync.RWMutex
	rooms  map[string]*Room
	logger *zap.Logger
}

func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		rooms:  make(map[string]*Room),
		logger: logger,
	}
}

func (r *Registry) GetOrCreate(ctx context.Context, roomID string, userID uint64) *Room {
	logger := r.logger.With(zap.String("roomID", roomID), zap.Uint64("userID", userID))

	r.mu.RLock()
	room, ok := r.rooms[roomID]
	r.mu.RUnlock()
	if ok {
		logger.Info("Room already exists")
		return room
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if room, ok = r.rooms[roomID]; ok {
		logger.Info("Room already exists")
		return room
	}

	room = NewRoom(roomID, r.deleteRoom, r.logger)
	r.rooms[roomID] = room

	logger.Info("Created room")

	return room
}

func (r *Registry) deleteRoom(roomID string) {
	r.mu.Lock()
	delete(r.rooms, roomID)
	r.mu.Unlock()

	r.logger.Info("Room has been emptied and deleted", zap.String("roomID", roomID))
}
