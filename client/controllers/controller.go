package controllers

import (
	"ChallengeFileServer/client/models"

	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	CHANNELS_FOLDER     = "/home/paugarrocho/channelsFolder"
	MAX_MEGABYTES       = 20
	MAX_BUFFER_CAPACITY = MAX_MEGABYTES * 1024 * 1024
)

const (
	ERR_CONNECTING_SERVER   = "Error connecting to server: %s"
	ERR_CREATING_FOLDER     = "Error creating client folder: %s"
	ERR_READING_CLIENT_DATA = "Error reading data from client: %s"
	ERR_PATH_NOT_EXISTS     = "The specified path does not exist"
	ERR_FILE_SIZE_LENGTH    = "The file size exceeds the maximum allowed of %d MB"
	ERR_READING_FILE        = "Error reading file: %s"
	MSG_GOODBYE             = "Goodbye!"
	MSG_CONNECTION_CLOSED   = "The server has closed the connection"
)

func GetFolder(conn net.Conn) string {
	addressData := strings.Split(conn.LocalAddr().String(), ":")
	return CHANNELS_FOLDER + "/" + addressData[1]
}

func CreateFolder(conn net.Conn) error {
	path := GetFolder(conn)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func CopyFile(fileName, fileData string, conn net.Conn) error {
	size := len(fileData)
	decoded := make([]byte, size*size/base64.StdEncoding.DecodedLen(size))
	_, err := base64.StdEncoding.Decode(decoded, []byte(fileData))
	if err != nil {
		return err
	}

	filePath := GetFolder(conn) + "/" + fileName
	fileCreate, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fileCreate.Close()

	if _, err := fileCreate.Write(decoded); err != nil {
		return err
	}

	if err := fileCreate.Sync(); err != nil {
		return err
	}

	return nil
}

func DecodeFile(filePath string) (models.File, string) {
	fileOpen, err := os.Open(filePath)
	if err != nil {
		return models.File{}, ERR_PATH_NOT_EXISTS
	}
	defer fileOpen.Close()

	reader := bufio.NewReaderSize(fileOpen, MAX_BUFFER_CAPACITY)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return models.File{}, err.Error()
	}

	if len(content) > MAX_BUFFER_CAPACITY {
		return models.File{}, fmt.Sprintf(ERR_FILE_SIZE_LENGTH, MAX_MEGABYTES)
	}

	pathData := strings.Split(filePath, "/")
	fileName := strings.Replace(pathData[len(pathData)-1], " ", "", -1)
	mimeType := http.DetectContentType(content)
	encoded := base64.StdEncoding.EncodeToString(content)

	file := models.File{
		Name: fileName,
		Type: mimeType,
		Data: encoded,
	}

	return file, ""
}
