package main

import (
	"net"
	"telecon/network"
	"telecon/utils"
)

type Client struct {
	net.Conn
	username string
	password string
	input    chan network.Packet
	output   chan network.Packet
	logged   bool
}

func (c *Client) Run() {
	for {
		select {
		case packet := <-c.input:
			go c.HandlePacket(packet)
		case packet := <-c.output:
			go c.SendPacket(packet)
		}
	}
}

func (c *Client) HandlePacket(packet network.Packet) {
	switch packet.GetType() {
	case network.PK_DISCONNECT:
		Print("You were disconnected from chat server: " + utils.BytesToStr(packet.Data[0]))
		Stop()
	case network.PK_MESSAGE:
		Print(utils.BytesToStr(packet.Data[0]))
	case network.PK_LOGIN:
		if c.logged {
			// Error
		} else {
			c.logged = true
		}
	default:
		log.Error("Unhandled packet! Dumping...")
		packet.Dump()
		Stop()
	}
}

func (c *Client) SendPacket(packet network.Packet) {
	packet.Put(c)
}

func (c *Client) Join() {
	pk := network.GetPacket(network.PK_LOGIN)
	pk.Data[0] = utils.StrToBytes(c.username)
	pk.Data[1] = utils.StrToBytes(c.password)
	c.SendPacket(*pk)
}
