package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"live-polling-backend/services"
)

type Hub struct {
	rooms        map[string]map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	redisService *services.RedisService
	mu           sync.RWMutex
	subscribers  map[string]context.CancelFunc
}

func NewHub(redisService *services.RedisService) *Hub {
	return &Hub{
		rooms:        make(map[string]map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		redisService: redisService,
		subscribers:  make(map[string]context.CancelFunc),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, exists := h.rooms[client.PollID]; !exists {
				h.rooms[client.PollID] = make(map[*Client]bool)
				// Start listening to Redis Pub/Sub for this poll
				h.startRedisSubscription(client.PollID)
			}
			h.rooms[client.PollID][client] = true
			log.Printf("[WebSocket Hub] Client joined poll room: %s (Total in room: %d)", client.PollID, len(h.rooms[client.PollID]))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, exists := h.rooms[client.PollID]; exists {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					log.Printf("[WebSocket Hub] Client left poll room: %s (Remaining: %d)", client.PollID, len(clients))
				}
				if len(clients) == 0 {
					delete(h.rooms, client.PollID)
					// Cancel Redis subscription if no clients left
					if cancel, hasSub := h.subscribers[client.PollID]; hasSub {
						cancel()
						delete(h.subscribers, client.PollID)
						log.Printf("[WebSocket Hub] Stopped Redis subscription for idle poll room: %s", client.PollID)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) startRedisSubscription(pollID string) {
	if h.redisService == nil {
		return
	}
	if _, alreadySubbed := h.subscribers[pollID]; alreadySubbed {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.subscribers[pollID] = cancel

	go func(pID string, subCtx context.Context) {
		pubsub := h.redisService.Subscribe(subCtx, pID)
		if pubsub == nil {
			return
		}
		defer pubsub.Close()

		ch := pubsub.Channel()
		log.Printf("[WebSocket Hub] Listening to Redis Pub/Sub messages for poll: %s", pID)

		for {
			select {
			case <-subCtx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				// Broadcast payload to all clients connected to this poll
				h.BroadcastToRoom(pID, []byte(msg.Payload))
			}
		}
	}(pollID, ctx)
}

func (h *Hub) BroadcastToRoom(pollID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, exists := h.rooms[pollID]
	if !exists {
		return
	}

	for client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(clients, client)
		}
	}
}

// SendInitialState directly sends snapshot to a newly connected client
func (h *Hub) SendInitialState(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err == nil {
		select {
		case client.Send <- data:
		default:
		}
	}
}
