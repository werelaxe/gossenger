package common

import "github.com/gorilla/websocket"

type UpgradeReminder map[uint]*websocket.Conn

func (um UpgradeReminder) Close() {
	for _, v := range um {
		v.Close()
	}
}
