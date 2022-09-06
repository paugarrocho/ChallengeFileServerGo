package controllers

import (
	"ChallengeFileServerGo/server/models"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	API_PORT   = 3000
	API_ORIGIN = "*"

	ERR_MESSAGE_IN          = "Error in %s"
	ERR_WRONG_COMM          = "Incorrect or incomplete command"
	ERR_UNDEF_MSG           = "Unspecified message"
	ERR_UNDEF_CHAN          = "Unspecified channel"
	ERR_UNDEF_PATH          = "Unspecified file"
	ERR_UNDEF_FILEDATA      = "File data not specified"
	ERROR_CANAL_INEXISTENTE = "Canal no encontrado"
	ERROR_SUBS              = "Tiene que estar subscripto al canal"
	ERR_NOT_FOUND_USER      = "User not found"
	ERR_DECODING_FILE       = "Error getting base64 data"
	MSG_CHANNEL_EXISTS      = "Channel already exists"
	MSG_FROM_CHANNEL        = ""
	MSJ_SUBS                = "Suscripción exitosa"
	MSJ_ARCHIVO_ENVIADO     = "Archivo enviado correctamente"
	MSJ_ARCHIVO_RECIBIDO    = "Recibió el archivo %s"
)

var UserMessages = make(chan models.Message)
var Messages = make(chan models.Message)
var Users []models.User
var Channels []models.ChannelRoom
var Files []models.File

func CreateMessage(msg string, user models.User) models.Message {
	return models.Message{
		Text:    msg,
		Address: user.Address,
		Channel: user.Channel,
	}
}

func CreateUser(conn net.Conn) {
	user := models.User{
		Address: conn.RemoteAddr().String(),
		Conn:    conn,
	}
	Users = append(Users, user)
}

func IsDestinationUser(message models.Message, user models.User) bool {
	return message.Address == user.Conn.RemoteAddr().String()
}

func IsUserInChannel(message models.Message, user models.User) bool {
	return message.Channel.Name == user.Channel.Name
}

func FindUserByAddress(address string) *models.User {
	for i := range Users {
		if Users[i].Address != address {
			continue
		}
		return &Users[i]
	}
	return nil
}

func DecodeCommand(command, address string) (string, string) {
	commandParts := strings.Split(command, " ")
	if len(commandParts) < 1 {
		return ERR_WRONG_COMM, ""
	}

	ownMessage, othersMessage := "", ""
	switch commandParts[0] {
	case "list":
		ownMessage = ListAllChannels(address)
	case "create":
		ownMessage = ERR_UNDEF_CHAN
		if commandParts[1] != "" {
			ownMessage = CreateChannel(commandParts[1:], address)
		}
	case "subscribe":
		ownMessage = ERR_UNDEF_CHAN
		if commandParts[1] != "" {
			ownMessage = SubscribeToChannel(commandParts[1:], address)
		}

	case "send":
		ownMessage = ERR_UNDEF_PATH
		if commandParts[1] != "" {
			ownMessage = ERR_UNDEF_FILEDATA
			if commandParts[2] != "" {
				ownMessage, othersMessage = SendFileToChannel(commandParts[1:], address)
			}
		}
	default:
		ownMessage = "Comando inválido"
	}

	return ownMessage, othersMessage
}

func JoinCommands(commands []string) string {
	return strings.Join(commands, " ")
}

func FindChannelByName(channelName string) *models.ChannelRoom {
	for i := range Channels {
		if Channels[i].Name != channelName {
			continue
		}
		return &Channels[i]
	}
	return nil
}

func CreateChannel(commands []string, address string) string {
	response := MSG_CHANNEL_EXISTS
	channelName := JoinCommands(commands)
	if channel := FindChannelByName(channelName); channel == nil {
		user := FindUserByAddress(address)
		if user == nil {
			log.Fatal(ERR_NOT_FOUND_USER)
		}

		newChannel := models.ChannelRoom{
			Name: channelName,
		}
		Channels = append(Channels, newChannel)
		user.Channel.Name = newChannel.Name
		response = "Canal creado correctamente"
	}
	return response
}

func ListAllChannels(address string) string {
	user := FindUserByAddress(address)
	if user == nil {
		log.Fatal(ERR_NOT_FOUND_USER)
	}

	response := []string{"Lista de Canales:"}
	for _, ch := range Channels {
		subscribed := ""
		if user.Channel.Name == ch.Name {
			subscribed = "<Suscripto>"
		}
		response = append(response, fmt.Sprintf("\t%s %s", ch.Name, subscribed))
	}
	return strings.Join(response, "\n")
}

func SubscribeToChannel(commands []string, address string) string {
	response := ERROR_CANAL_INEXISTENTE
	channelName := JoinCommands(commands)
	if channel := FindChannelByName(channelName); channel != nil {
		user := FindUserByAddress(address)
		if user == nil {
			log.Fatal(ERR_NOT_FOUND_USER)
		}
		user.Channel.Name = channel.Name
		response = MSJ_SUBS
	}
	return response
}

func SendFileToChannel(commands []string, address string) (string, string) {
	user := FindUserByAddress(address)
	if user == nil {
		log.Fatal(ERR_NOT_FOUND_USER)
	}

	responseOwn := ERROR_SUBS
	responseOthers := ""
	if user.Channel.Name != "" {
		responseOwn = ERROR_CANAL_INEXISTENTE
		if channel := FindChannelByName(user.Channel.Name); channel != nil {
			file, err := CreateBase64File(*user, commands)
			if err != nil {
				return ERR_DECODING_FILE, ""
			}

			responseOwn = MSJ_ARCHIVO_ENVIADO
			responseOthers = fmt.Sprintf(MSJ_ARCHIVO_RECIBIDO, file.Name) + "~" + file.Name + "~" + file.Data
		}
	}
	return responseOwn, responseOthers
}

func CreateBase64File(user models.User, commands []string) (models.File, error) {
	//URI scheme: data:[<media type>][;base64],<data> (default is text/plain;charset=US-ASCII)
	dataFile := strings.Split(commands[1], ",")
	mediaType := "text/plain"
	data := dataFile[0]

	if len(dataFile) > 1 {
		dataType := strings.Split(dataFile[0], ";")
		mediaType = dataType[0][strings.IndexByte(dataType[0], ':')+1:]
		data = dataFile[1]
	}

	//Renombra el archivo si ya existe
	count := 1
	fileName := commands[0]
	fileValid := false
	for !fileValid {
		fileOldName := fileName
		for _, f := range Files {
			if f.Name != fileName {
				continue
			}

			name := strings.Split(fileName, ".")
			fileName = fmt.Sprintf("%s-%d.%s", name[0], count, name[1])
			count = count + 1
		}

		if fileOldName == fileName {
			fileValid = true
		}
	}

	file := models.File{
		Name:    fileName,
		Type:    mediaType,
		Data:    data,
		Channel: user.Channel,
	}
	Files = append(Files, file)
	return file, nil
}
