package controllers

import (
	"ChallengeFileServerGo/client/models"

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
	FOLDER     = "/home/paugarrocho/Canales"
	MAX_MB     = 20
	MAX_BUFFER = MAX_MB * 1024 * 1024
)

const (
	ERROR_CONEXION         = "Error al conectarse al servidor: %s"
	ERROR_CREACION_CARPETA = "Error creando carpeta: %s"
	ERROR_LECTURA_CLIENTE  = "Error leyendo los datos del cliente: %s"
	ERROR_RUTA_INEXISTENTE = "La ruta especificada no existe"
	ERROR_TAM_ARCH         = "El tamaño del archivo excede el máximo permitido %d MB"
	ERROR_LECTURA_ARCHIVO  = "Error de lectura del archivo: %s"
	MSJ_SALIDA             = "Conexión terminada"
	MSJ_CONEXION_CERRADA   = "El servidor ha terminado la conexión"
)

func GetFolder(conn net.Conn) string {
	addressData := strings.Split(conn.LocalAddr().String(), ":")
	return FOLDER + "/" + addressData[1]
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
		return models.File{}, ERROR_RUTA_INEXISTENTE
	}
	defer fileOpen.Close()

	reader := bufio.NewReaderSize(fileOpen, MAX_BUFFER)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return models.File{}, err.Error()
	}

	if len(content) > MAX_BUFFER {
		return models.File{}, fmt.Sprintf(ERROR_TAM_ARCH, MAX_MB)
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
