package common

import "github.com/gorilla/websocket"

const (
	MessagesConnType = "Messages"
	ChatsConnType    = "Chats"
)

type ConnectionKeeper map[string]map[uint]*websocket.Conn

func (connKeeper ConnectionKeeper) Init() {
	connKeeper[MessagesConnType] = make(map[uint]*websocket.Conn)
	connKeeper[ChatsConnType] = make(map[uint]*websocket.Conn)
}

func (connKeeper ConnectionKeeper) Close() {
	for _, conn := range connKeeper {
		for _, conn := range conn {
			conn.Close()
		}
	}
}
