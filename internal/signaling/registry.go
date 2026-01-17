package signaling

import (
	"chatter/pkg/logger"
	"sync"
)

type Registry struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewRegistry() *Registry {
	return &Registry{
		rooms: make(map[string]*Room),
	}
}

func (r *Registry) GetOrCreate(roomID string) *Room {
	r.mu.RLock()
	room, ok := r.rooms[roomID]
	r.mu.RUnlock()
	if ok {
		return room
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if room, ok = r.rooms[roomID]; ok {
		return room
	}

	room = NewRoom(roomID, r.deleteRoom, logger.Logger)
	r.rooms[roomID] = room
	return room
}

func (r *Registry) deleteRoom(roomID string) {
	r.mu.Lock()
	delete(r.rooms, roomID)
	r.mu.Unlock()
}
