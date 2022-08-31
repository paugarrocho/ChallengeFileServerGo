package main

import (
	"ChallengeFileServerGo/server/controllers"
	"ChallengeFileServerGo/server/router"

	"bufio"
	"fmt"
	"log"
	"net"
)

const (
	CONN_HOST           = "localhost"
	CONN_PORT           = "7777"
	CONN_TYPE           = "tcp"
	MAX_MEGABYTES       = 20
	MAX_BUFFER_CAPACITY = MAX_MEGABYTES * 1024 * 1024
)

func main() {
	listen, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		log.Fatal(fmt.Sprintf(controllers.ERR_MESSAGE_IN, "net.Listen:"), err.Error())
	}

	/*controllers.CreateDefaultChannels()*/
	go broadcaster()
	go apiServer()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(fmt.Sprintf(controllers.ERR_MESSAGE_IN, "listen.Accept:"), err.Error())
			continue
		}
		defer conn.Close()

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	controllers.CreateUser(conn)
	address := conn.RemoteAddr().String()

	//Get the data sent by a client
	input := bufio.NewScanner(conn)
	buf := make([]byte, MAX_BUFFER_CAPACITY)
	input.Buffer(buf, MAX_BUFFER_CAPACITY)

	for input.Scan() {
		if input.Text() == "" {
			continue
		}

		user := controllers.FindUserByAddress(address)
		if user == nil {
			log.Fatal(controllers.ERR_NOT_FOUND_USER)
		}

		messageToOwnUser, messageToOtherUsers := controllers.DecodeCommand(input.Text(), user.Address)
		if messageToOwnUser != "" {
			//Send message to same client
			controllers.UserMessages <- controllers.CreateMessage(messageToOwnUser, *user)
		}
		if messageToOtherUsers != "" {
			//Send message to another clients in the channel
			controllers.Messages <- controllers.CreateMessage(messageToOtherUsers, *user)
		}
	}

	controllers.DeleteUser(address)
}

func broadcaster() {
	//"|" indicates the end of the message
	for {
		select {
		case message := <-controllers.UserMessages:
			//Messages for same client (ex: list)
			for _, user := range controllers.Users {
				if !controllers.IsDestinationUser(message, user) {
					continue
				}
				fmt.Fprintln(user.Conn, "$$ "+message.Text+"|")
			}
		case message := <-controllers.Messages:
			for _, user := range controllers.Users {
				if controllers.IsDestinationUser(message, user) {
					continue
				}

				//Send the message to another clients in the same channel (ex: send file)
				if controllers.IsUserInChannel(message, user) {
					fmt.Fprintln(user.Conn, "\n$$ "+controllers.MSG_FROM_CHANNEL+": "+message.Text+"|")
				}
			}
		}
	}
}

func apiServer() {
	router.SetRoutes()
}
