package main

import (
	"bytes"
	"flag"
	"net"
	"os"
	"telecon/logger"
	"telecon/network"
	"telecon/utils"
	"time"
)

const (
	VERSION  = "2.0.0"
	CODENAME = "Elite"
	BUILD    = "1"
)

var client Client

var log logger.Logger = logger.Logger{}

func main() {
	flag.Parse()
	log.Info("Starting Telecon v" + VERSION + " Bx" + BUILD + " [" + CODENAME + "]")

	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Critical(err)
	}

	client = Client{
		conn,
		"Primus",
		"SomePassword",
		make(chan network.Packet),
		make(chan network.Packet),
		false,
	}

	client.Join()

	go client.Run()

	go func() {
		st := time.Now().Unix()
		var idle bool = false
		for {
			if (time.Now().Unix() - st) > 5 {
				idle = true
				break
			}
		}
		if idle {
			log.Info("Connection was idle state")
			Stop()
			return
		}
	}()

	// read from connection
	for {
		var buffer bytes.Buffer
		var current []byte = make([]byte, 1024*5)
		for {
			n, err := conn.Read(current)
			if err != nil {
				if err.Error() == "EOF" {
					pk := network.GetPacket(network.PK_DISCONNECT)
					pk.Data[0] = utils.StrToBytes("Server stopped")
					client.input <- *pk
					break
				}
			}
			buffer.Write(current[:n])
			packets, rest := network.ReadPackets(buffer.Bytes())
			buffer.Reset()
			buffer.Write(rest)
			for _, p := range packets {
				p.Dump()
				client.input <- p
			}
		}
	}

	Stop()
}

func Stop() {
	log.Info("Stopping...")
	log.Info("Stopped.")
	os.Exit(0)
}

func Print(message string) {
	logger.Log(logger.BLANK, "", message)
}
