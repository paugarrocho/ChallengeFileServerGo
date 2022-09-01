package main

import (
	"ChallengeFileServerGo/client/controllers"

	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	HOST = "localhost"
	PORT = "7777"
	TIPO = "tcp"
)

func main() {
	conn, err := net.Dial(TIPO, HOST+":"+PORT)
	if err != nil {
		log.Fatalf(fmt.Sprintf(controllers.ERROR_CONEXION, err.Error()))
		return
	}
	defer conn.Close()

	if err := controllers.CreateFolder(conn); err != nil {
		log.Fatalf(fmt.Sprintf(controllers.ERROR_CREACION_CARPETA, err.Error()))
		return
	}

	receiveFromServer := make(chan string)

	for {
		go handleReceiveFromServer(conn, receiveFromServer)
		go handleSendToServer(conn)

		res := <-receiveFromServer
		//Prints the server response
		fmt.Println(res)

	}
}

func handleReceiveFromServer(conn net.Conn, chIn chan string) {
	//"|" indicates the end of the message
	message, err := bufio.NewReaderSize(conn, controllers.MAX_BUFFER).ReadString('|')
	if err != nil {
		exitClient(false)
	}

	//If the message contains "~", it means that the information of a file is coming
	message = strings.Replace(message, "|", "", 1)
	if strings.Contains(message, "~") {
		fileData := strings.Split(message, "~")
		message = fileData[0]

		err := controllers.CopyFile(fileData[1], fileData[2], conn)
		if err != nil {
			message = err.Error()
		}
	}

	chIn <- message
}

func handleSendToServer(conn net.Conn) {
	reader := bufio.NewReaderSize(os.Stdin, controllers.MAX_BUFFER)
	fmt.Print(">> ")
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(fmt.Sprintf(controllers.ERROR_LECTURA_CLIENTE, err.Error()))
	}

	text = strings.TrimSpace(text)
	if text != "" {
		if text == "exit" {
			exitClient(true)
		}

		if commandParts := strings.Split(text, " "); commandParts[0] == "send" {
			filePath := strings.Join(commandParts[1:], " ")
			file, err := controllers.DecodeFile(filePath)
			text = commandParts[0] + " " + file.Name + " " + file.Data
			if err != "" {
				text = "image-wrcomm " + fmt.Sprintf(controllers.ERROR_LECTURA_ARCHIVO, err)
			}
		}
	} else {
		text = "wrcomm"
	}

	fmt.Fprintf(conn, text+"\n")
}

func exitClient(userExit bool) {
	message := "\n$$ " + fmt.Sprintf(controllers.MSJ_CONEXION_CERRADA)
	if userExit {
		message = "$$ " + fmt.Sprintf(controllers.MSJ_SALIDA)
	}
	fmt.Println(message)
	os.Exit(0)
}
