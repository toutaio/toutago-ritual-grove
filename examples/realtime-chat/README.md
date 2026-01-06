# Real-time Chat Example

This example demonstrates building a real-time chat application with Toutā, Inertia.js, and WebSockets.

## Features

- 💬 **Real-time messaging** with WebSockets
- 👥 **Multiple chat rooms**
- ✍️ **Typing indicators**
- 📎 **File attachments**
- 😀 **Emoji support**
- 📱 **Mobile responsive**
- 🔔 **Push notifications**
- 👤 **User presence** (online/offline status)
- 🔍 **Message search**
- 📌 **Pinned messages**

## Quick Start

```bash
# Generate base project
touta ritual init blog

# Choose:
# - Frontend: inertia-vue
# - Enable SSR: no (not needed for chat)
# - Real-time: websockets
```

## Architecture

```
realtime-chat/
├── handlers/
│   ├── chat/
│   │   ├── room.go           # Room management
│   │   ├── message.go        # Message handling
│   │   └── ws.go             # WebSocket handler
│   └── middleware/
│       └── auth.go
├── services/
│   ├── chat_service.go       # Business logic
│   └── presence_service.go   # User presence tracking
├── models/
│   ├── room.go
│   ├── message.go
│   └── user.go
├── resources/js/
│   ├── Pages/
│   │   ├── Chat/
│   │   │   ├── Index.vue     # Room list
│   │   │   └── Room.vue      # Chat room
│   │   └── Auth/
│   │       └── Login.vue
│   ├── Components/
│   │   ├── MessageList.vue
│   │   ├── MessageInput.vue
│   │   ├── UserList.vue
│   │   └── TypingIndicator.vue
│   └── composables/
│       ├── useWebSocket.js   # WebSocket composable
│       └── useChat.js        # Chat logic
└── public/
    └── sounds/
        └── notification.mp3
```

## Key Implementation

### WebSocket Handler (Go)

```go
// handlers/chat/ws.go
package chat

import (
    "github.com/gorilla/websocket"
    "github.com/toutaio/toutago-cosan-router"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Configure properly for production
    },
}

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan *Message
    register   chan *Client
    unregister chan *Client
    rooms      map[string]*Room
}

type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    user     *models.User
    roomID   string
}

type Message struct {
    Type    string      `json:"type"`
    RoomID  string      `json:"room_id"`
    UserID  int         `json:"user_id"`
    User    string      `json:"user"`
    Content string      `json:"content"`
    Data    interface{} `json:"data,omitempty"`
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan *Message),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        rooms:      make(map[string]*Room),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            h.notifyPresence(client.roomID, client.user, "online")
            
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
                h.notifyPresence(client.roomID, client.user, "offline")
            }
            
        case message := <-h.broadcast:
            h.broadcastToRoom(message.RoomID, message)
        }
    }
}

func (h *Hub) broadcastToRoom(roomID string, message *Message) {
    for client := range h.clients {
        if client.roomID == roomID {
            select {
            case client.send <- []byte(message.ToJSON()):
            default:
                close(client.send)
                delete(h.clients, client)
            }
        }
    }
}

func (h *WSHandler) ServeWS(c *cosan.Context) error {
    user := c.Get("user").(*models.User)
    roomID := c.Param("roomID")
    
    conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
    if err != nil {
        return err
    }
    
    client := &Client{
        hub:    h.hub,
        conn:   conn,
        send:   make(chan []byte, 256),
        user:   user,
        roomID: roomID,
    }
    
    h.hub.register <- client
    
    go client.readPump()
    go client.writePump()
    
    return nil
}

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    
    for {
        var msg Message
        err := c.conn.ReadJSON(&msg)
        if err != nil {
            break
        }
        
        msg.UserID = c.user.ID
        msg.User = c.user.Name
        msg.RoomID = c.roomID
        
        // Save to database
        c.hub.saveMessage(&msg)
        
        // Broadcast to room
        c.hub.broadcast <- &msg
    }
}

func (c *Client) writePump() {
    defer c.conn.Close()
    
    for message := range c.send {
        if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
            break
        }
    }
}
```

### Vue WebSocket Composable

```javascript
// composables/useWebSocket.js
import { ref, onMounted, onUnmounted } from 'vue'

export function useWebSocket(roomID) {
  const ws = ref(null)
  const messages = ref([])
  const connected = ref(false)
  const typing = ref([])
  
  const connect = () => {
    ws.value = new WebSocket(`ws://localhost:3000/ws/chat/${roomID}`)
    
    ws.value.onopen = () => {
      connected.value = true
      console.log('WebSocket connected')
    }
    
    ws.value.onmessage = (event) => {
      const message = JSON.parse(event.data)
      
      switch (message.type) {
        case 'message':
          messages.value.push(message)
          playNotificationSound()
          break
        case 'typing':
          handleTyping(message)
          break
        case 'presence':
          handlePresence(message)
          break
      }
    }
    
    ws.value.onclose = () => {
      connected.value = false
      console.log('WebSocket disconnected')
      // Reconnect after 3 seconds
      setTimeout(connect, 3000)
    }
    
    ws.value.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }
  
  const send = (message) => {
    if (ws.value && connected.value) {
      ws.value.send(JSON.stringify(message))
    }
  }
  
  const sendMessage = (content) => {
    send({
      type: 'message',
      content,
      timestamp: new Date().toISOString()
    })
  }
  
  const sendTyping = (isTyping) => {
    send({
      type: 'typing',
      data: { isTyping }
    })
  }
  
  const handleTyping = (message) => {
    if (message.data.isTyping) {
      if (!typing.value.includes(message.user)) {
        typing.value.push(message.user)
      }
    } else {
      typing.value = typing.value.filter(u => u !== message.user)
    }
    
    // Clear typing after 3 seconds
    setTimeout(() => {
      typing.value = typing.value.filter(u => u !== message.user)
    }, 3000)
  }
  
  const handlePresence = (message) => {
    // Update user presence
    console.log(`${message.user} is ${message.data.status}`)
  }
  
  const playNotificationSound = () => {
    const audio = new Audio('/sounds/notification.mp3')
    audio.play().catch(() => {})
  }
  
  onMounted(connect)
  onUnmounted(() => {
    if (ws.value) {
      ws.value.close()
    }
  })
  
  return {
    messages,
    connected,
    typing,
    sendMessage,
    sendTyping
  }
}
```

### Chat Room Component

```vue
<!-- Pages/Chat/Room.vue -->
<template>
  <div class="flex h-screen">
    <!-- Sidebar -->
    <div class="w-64 bg-gray-800 text-white">
      <div class="p-4 border-b border-gray-700">
        <h2 class="text-xl font-bold">{{ room.name }}</h2>
      </div>
      <UserList :users="onlineUsers" />
    </div>
    
    <!-- Main Chat -->
    <div class="flex-1 flex flex-col">
      <!-- Messages -->
      <div ref="messageContainer" class="flex-1 overflow-y-auto p-4">
        <MessageList :messages="messages" />
        <TypingIndicator v-if="typing.length > 0" :users="typing" />
      </div>
      
      <!-- Input -->
      <div class="border-t p-4">
        <MessageInput
          v-model="messageText"
          @send="handleSend"
          @typing="handleTyping"
          :disabled="!connected"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import { useWebSocket } from '@/composables/useWebSocket'
import MessageList from '@/Components/MessageList.vue'
import MessageInput from '@/Components/MessageInput.vue'
import TypingIndicator from '@/Components/TypingIndicator.vue'
import UserList from '@/Components/UserList.vue'

const props = defineProps({
  room: Object,
  onlineUsers: Array,
  initialMessages: Array
})

const messageText = ref('')
const messageContainer = ref(null)

const { messages, connected, typing, sendMessage, sendTyping } = 
  useWebSocket(props.room.id)

// Initialize with server-rendered messages
messages.value = props.initialMessages

const handleSend = () => {
  if (messageText.value.trim()) {
    sendMessage(messageText.value)
    messageText.value = ''
  }
}

let typingTimeout
const handleTyping = (isTyping) => {
  clearTimeout(typingTimeout)
  
  if (isTyping) {
    sendTyping(true)
    typingTimeout = setTimeout(() => {
      sendTyping(false)
    }, 1000)
  } else {
    sendTyping(false)
  }
}

// Auto-scroll to bottom on new messages
watch(messages, async () => {
  await nextTick()
  messageContainer.value.scrollTop = messageContainer.value.scrollHeight
})
</script>
```

### Message Input with Emoji

```vue
<!-- Components/MessageInput.vue -->
<template>
  <div class="relative">
    <div class="flex items-center gap-2">
      <button
        @click="showEmojiPicker = !showEmojiPicker"
        class="p-2 hover:bg-gray-100 rounded"
      >
        😀
      </button>
      
      <input
        v-model="localValue"
        type="text"
        placeholder="Type a message..."
        @keyup.enter="$emit('send')"
        @input="handleInput"
        class="flex-1 p-2 border rounded"
        :disabled="disabled"
      />
      
      <button
        @click="$emit('send')"
        :disabled="disabled || !localValue.trim()"
        class="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
      >
        Send
      </button>
    </div>
    
    <EmojiPicker
      v-if="showEmojiPicker"
      @select="insertEmoji"
      class="absolute bottom-full mb-2"
    />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import EmojiPicker from './EmojiPicker.vue'

const props = defineProps({
  modelValue: String,
  disabled: Boolean
})

const emit = defineEmits(['update:modelValue', 'send', 'typing'])

const localValue = ref(props.modelValue)
const showEmojiPicker = ref(false)

watch(() => props.modelValue, (val) => {
  localValue.value = val
})

let typingTimer
const handleInput = () => {
  emit('update:modelValue', localValue.value)
  
  clearTimeout(typingTimer)
  emit('typing', true)
  
  typingTimer = setTimeout(() => {
    emit('typing', false)
  }, 1000)
}

const insertEmoji = (emoji) => {
  localValue.value += emoji
  emit('update:modelValue', localValue.value)
  showEmojiPicker.value = false
}
</script>
```

## Running the Application

### 1. Install dependencies

```bash
go mod tidy
npm install
```

### 2. Start the server

```bash
# Terminal 1: Go server
go run main.go

# Terminal 2: Frontend dev
npm run dev
```

### 3. Open multiple browsers

Test real-time features by opening http://localhost:3000/chat in multiple browser windows.

## Features in Detail

### Typing Indicators

Users see when others are typing in real-time.

### User Presence

Online/offline status updated automatically via WebSocket heartbeat.

### Message Persistence

All messages saved to database, loaded on room join.

### File Uploads

```go
func (h *MessageHandler) Upload(c *cosan.Context) error {
    file, err := c.FormFile("file")
    if err != nil {
        return err
    }
    
    // Save file
    path := h.storage.Save(file)
    
    // Broadcast file message
    h.hub.broadcast <- &Message{
        Type: "message",
        RoomID: c.Param("roomID"),
        Content: "",
        Data: map[string]string{
            "type": "file",
            "url": path,
            "filename": file.Filename,
        },
    }
    
    return c.JSON(200, map[string]string{"url": path})
}
```

## Production Considerations

- Use Redis for pub/sub across multiple server instances
- Implement rate limiting for messages
- Add message moderation/filtering
- Enable HTTPS for secure WebSockets (wss://)
- Add user blocking/reporting features
- Implement message read receipts
- Add voice/video call support

## License

MIT
