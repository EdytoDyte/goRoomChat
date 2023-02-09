package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type keys struct {
	Protocol []byte
	Publick  []byte
}
type msges struct {
	Protocol []byte
	Mensaje  []byte
}

var privateKey *rsa.PrivateKey //<-- private key from the client
var publicKey *rsa.PublicKey   //<-- public key from the server
var publicKeySD *rsa.PublicKey //<-- public key from the client
var NameRoom string
var key bool //<--- Check if we got the key from the server

// The code is a chat client that communicates with a chat server using the TCP protocol. It allows the user to either create a chat room or join an existing one. The chat messages are encrypted using RSA encryption.
func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	conns := make(map[net.Conn]string)
	go handleIncomingMessages(conn, conns)
	getHash(conn)
	fmt.Println("::: Welcome to the Go Room Chat! :::")
	fmt.Println("_______________________________")
	fmt.Println("|          Options            |")
	fmt.Println("|_____________________________|")
	fmt.Println("|1. Create an chat room       |")
	fmt.Println("|2. Join a chat room          |")
	fmt.Println("|3. Exit the goRoomChat       |")
	fmt.Println("|_____________________________|")
	fmt.Println(":::   Select your option   :::")
	option(conn)
	getUser(conn)

}

// Handles incoming messages from the server. It reads the messages, unmarshals the JSON, decrypts the messages using RSA encryption, and prints them to the console.
func handleIncomingMessages(conexion net.Conn, conexiones map[net.Conn]string) {
	defer conexion.Close()
	hash, _ := bufio.NewReader(conexion).ReadString('\n')
	var Clavese keys
	err := json.Unmarshal([]byte(hash), &Clavese)
	if err != nil {
		fmt.Println(err)
	}
	if string(Clavese.Protocol) == "key" {
		pubkey, _ := x509.ParsePKIXPublicKey(Clavese.Publick)
		publicKey = pubkey.(*rsa.PublicKey)
		fmt.Print("::: We recived the key from the server :::\n")
		if publicKey != nil {
			key = true

			getMessages(conexion)

		}
	}
}

// Allows the user to input their username and send it to the server.
func getUser(conn net.Conn) {
	mssagedPrinted := false
	for {
		if key {
			fmt.Print("::: Introduce your username  :::\n")
			username, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			username = strings.TrimRight(username, "\n")
			username = strings.TrimSpace(username)
			if username != "" {
				conn.Write([]byte(username))
				conn.Write([]byte("\n"))

				IniGu(conn)

			} else {
				fmt.Print("::: The username cannot be blank :::\n")
				getUser(conn)
			}
			break
		} else {
			if !mssagedPrinted {
				fmt.Print("::: Waiting key from the server :::\n")
				mssagedPrinted = true

			}

		}
	}

}

// Prompts the user to select an option: create a room, join a room, or exit the app. The user's selection is sent to the server.
func option(conn net.Conn) {
	opcion, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	opcion = strings.TrimRight(opcion, "\n")
	opcion = strings.TrimSpace(opcion)
	switch opcion {
	case "1":
		conn.Write([]byte("Creating the room\n"))
		nombreSala(conn)
		fmt.Print("::: Creating room :::\n")
	case "2":
		conn.Write([]byte("Joining a room\n"))
		nombreSala(conn)
		fmt.Print("::: Joining a room :::\n")
	case "3":
		fmt.Println("You chose to exit the app.")
		os.Exit(0)
	default:
		fmt.Println("Invalid option.")
		option(conn)
	}
}

// Prompts the user to enter the name of the chat room they want to join or create.
func nombreSala(conn net.Conn) {
	fmt.Println("::: Enter room name  :::")
	nombresala, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	nombresala = strings.TrimRight(nombresala, "\n")
	nombresala = strings.TrimSpace(nombresala)
	if nombresala != "" {
		NameRoom = nombresala

		conn.Write([]byte(nombresala + "\n"))
	} else {
		fmt.Print("::: The room name cannot be blank :::\n")
		nombreSala(conn)
	}
}

// Generates the publick key and private
func getHash(conn net.Conn) {
	fmt.Println("::: Sending security key:::")
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return
	}
	privateKey = privatekey
	publicKeySD = &privatekey.PublicKey
	pemKey := x509.MarshalPKCS1PublicKey(publicKeySD)
	keyss := keys{
		Publick: pemKey,
	}
	public, _ := json.Marshal(keyss)
	conn.Write(public)
	conn.Write([]byte("\n"))
}

// Encrypts data using RSA encryption.
func encriptar(msg []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	mensaje, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, msg, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return mensaje, nil
}

// Decrypts data using RSA encryption.
func desencriptar(msg []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	mensaje, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, msg, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return mensaje, nil
}
func getMessages(conexion net.Conn) {

	for {
		mensajes, err := bufio.NewReader(conexion).ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed")
			return
		}
		Mensaje := string(mensajes)
		var message msges
		err2 := json.Unmarshal([]byte(Mensaje), &message)
		if string(message.Protocol) == "msg" {
			if err2 != nil {
				fmt.Print(err2)
			}
			mesDesen, _ := desencriptar([]byte(message.Mensaje), privateKey)
			updateView(GoCui, mesDesen)
		}
	}
}
