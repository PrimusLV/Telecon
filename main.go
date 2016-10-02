package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"net"
	"os"
	"telecon/input"
	"telecon/logger"
	"telecon/network"
	"telecon/utils"
	"time"
)

const (
	VERSION  = "2.0.0"
	CODENAME = "Elite"
	BUILD    = "3"
)

var client Client

var log logger.Logger = logger.Logger{}

var username *string = flag.String("username", "changeme", "Your username")
var password *string = flag.String("password", "", "In case you are an admin on target server, you have to provide a password to authenticate or else you'll be kicked")
var server *string = flag.String("server", "localhost:9000", "Target chat server")

func main() {
	flag.Parse()

	if *username == "changeme" {
		log.Info("Enter your username using `-username` flag")
		Stop()
	}

	log.Info("Starting Telecon-Client v" + VERSION + " Bx" + BUILD + " [" + CODENAME + "]")
	cf := color.New(color.FgGreen).SprintFunc()
	log.Info(fmt.Sprintf("Connecting to %s as %s", cf(*server), cf(*username)))
	conn, err := net.Dial("tcp", *server)
	if err != nil {
		log.Critical(err)
	}

	client = Client{
		conn,
		*username,
		*password,
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
			if client.logged {
				break
			}
		}
		if idle {
			log.Info("Connection was idle state")
			Stop()
			return
		}
	}()

	input.SetTarget(client)
	go input.Start()
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
