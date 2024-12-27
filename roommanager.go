package main

import (
	"fmt"
	"sync"
)

type RoomManager struct {
	rooms      map[string]*Hub
	createRoom chan string
	deleteRoom chan string
	mu         sync.Mutex
}

func newRoomManager() *RoomManager {
	return &RoomManager{
		rooms:      make(map[string]*Hub),
		createRoom: make(chan string),
		deleteRoom: make(chan string),
	}
}

func (rm *RoomManager) run() {
	for {
		select {
		case roomName := <-rm.createRoom:
			rm.mu.Lock()
			if _, exists := rm.rooms[roomName]; !exists {
				hub := newHub()
				rm.rooms[roomName] = hub
				go hub.run()
			}
			rm.mu.Unlock()
		case roomName := <-rm.deleteRoom:
			rm.mu.Lock()
			if hub, exists := rm.rooms[roomName]; exists {
				close(hub.broadcast)
				delete(rm.rooms, roomName)
			}
			rm.mu.Unlock()
		}
	}
}

func (rm *RoomManager) getRooms() []string {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rooms := make([]string, 0, len(rm.rooms))
	for name := range rm.rooms {
		rooms = append(rooms, name)
	}
	return rooms
}

func (rm *RoomManager) getRoom(roomName string) (*Hub, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	hub, ok := rm.rooms[roomName]
	if !ok {
		return nil, fmt.Errorf("room %s does not exist", roomName)
	}

	return hub, nil
}
