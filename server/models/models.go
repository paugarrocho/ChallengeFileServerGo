package models

import "net"

type Message struct {
	Text    string
	Address string
	Channel ChannelRoom
}

type ChannelRoom struct {
	Name string `json:"name"`
}

type User struct {
	Address string      `json:"address"`
	Conn    net.Conn    `json:"-"`
	Channel ChannelRoom `json:"channel"`
}

type File struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Data    string      `json:"data"`
	Channel ChannelRoom `json:"channel"`
}
