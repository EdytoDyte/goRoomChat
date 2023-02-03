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
	"time"
)

type keys struct {
	Publick []byte
}
type msges struct {
	Mensaje []byte
}

var privateKey *rsa.PrivateKey //<-- private key from the client
var publicKey *rsa.PublicKey   //<-- public key from the server
var publicKeySD *rsa.PublicKey //<-- public key from the client

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
	fmt.Println("::: You can send messages now :::")
	mensajes(conn)
}

// Handles incoming messages from the server. It reads the messages, unmarshals the JSON, decrypts the messages using RSA encryption, and prints them to the console.
func handleIncomingMessages(conexion net.Conn, conexiones map[net.Conn]string) {
	defer conexion.Close()
	hash, _ := bufio.NewReader(conexion).ReadString('\n')
	var Clavese keys
	err := json.Unmarshal([]byte(hash), &Clavese)
	if err != nil {
		fmt.Println(err)
		return
	}
	pubkey, _ := x509.ParsePKIXPublicKey(Clavese.Publick)
	publicKey = pubkey.(*rsa.PublicKey)
	fmt.Print("::: We recived the key from the server :::\n")
	for {
		mensajes, err := bufio.NewReader(conexion).ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed")
			return
		}
		var message msges
		err2 := json.Unmarshal([]byte(mensajes), &message)
		if err2 != nil {
			fmt.Println(err)
			return
		}
		mesDesen, _ := desencriptar([]byte(message.Mensaje), privateKey)
		fmt.Print(string(mesDesen))
	}
}

// Allows the user to input their username and send it to the server.
func getUser(conn net.Conn) {
	time.Sleep(2 * time.Second)
	fmt.Println("::: Introduce tu nombre de usuario  :::")
	username, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	conn.Write([]byte(username))
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
	default:
		fmt.Println("Invalid option.")
		option(conn)
	}
}

// Prompts the user to enter the name of the chat room they want to join or create.
func nombreSala(conn net.Conn) {
	fmt.Println("::: Enter room name  :::")
	nombreSala, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	nombreSala = strings.TrimRight(nombreSala, "\n")
	nombreSala = strings.TrimSpace(nombreSala)
	conn.Write([]byte(nombreSala + "\n"))
}

// Handles sending messages from the user to the server. The messages are encrypted using RSA encryption before being sent.
func mensajes(conn net.Conn) {
	for {
		mensajes, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		mensajes = strings.TrimRight(mensajes, "\n")
		mensajes = strings.TrimSpace(mensajes)
		msgEn, _ := encriptar([]byte(mensajes), publicKey)
		message := msges{
			Mensaje: msgEn,
		}
		msgJson, _ := json.Marshal(message)
		conn.Write(msgJson)
		conn.Write([]byte("\n"))
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
