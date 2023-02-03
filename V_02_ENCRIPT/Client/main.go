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

var privateKey *rsa.PrivateKey //<-- private key del cliente
var publicKey *rsa.PublicKey   //<-- public key del servidor
var publicKeySD *rsa.PublicKey //<-- public key del cliente
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
	fmt.Println(":::Bienvenido al chat de Go!:::")
	fmt.Println("_______________________________")
	fmt.Println("|          Opciones           |")
	fmt.Println("|_____________________________|")
	fmt.Println("|1. Crear un sala de chat     |")
	fmt.Println("|2. Meterse a una sala de chat|")
	fmt.Println("|3. Salir del AppChat de Go   |")
	fmt.Println("|_____________________________|")
	fmt.Println(":::   Introduce tu opcion   :::")
	option(conn)
	getUser(conn)
	fmt.Println("::: Ya puede enviar mensajes :::")
	mensajes(conn)
}
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
	fmt.Print("::: Se ha recibido la clave del servidor :::\n")
	for {
		mensajes, err := bufio.NewReader(conexion).ReadString('\n')
		if err != nil {
			fmt.Println("Conexión cerrada")
			return
		}
		var message msges                                  // Inicializar una variable que sera el contenedor de json
		err2 := json.Unmarshal([]byte(mensajes), &message) // Parsear el json para poder leerlos
		if err2 != nil {
			fmt.Println(err)
			return
		}
		mesDesen, _ := desencriptar([]byte(message.Mensaje), privateKey)
		fmt.Print(string(mesDesen))
	}
}
func getUser(conn net.Conn) {
	time.Sleep(2 * time.Second)
	fmt.Println("::: Introduce tu nombre de usuario  :::")
	// Nombre de usuario
	username, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	conn.Write([]byte(username))
}
func option(conn net.Conn) {
	opcion, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	opcion = strings.TrimRight(opcion, "\n")
	opcion = strings.TrimSpace(opcion)
	switch opcion {
	case "1":
		conn.Write([]byte("Crear sala\n"))
		nombreSala(conn)
		fmt.Print("::: Creando la sala :::\n")
	case "2":
		conn.Write([]byte("Unirse a una sala\n"))
		nombreSala(conn)
		fmt.Print("::: Uniendose a la sala :::\n")
	case "3":
		fmt.Println("Elegiste salir de la app.")
	default:
		fmt.Println("Opción inválida.")
		option(conn)
	}
}
func nombreSala(conn net.Conn) {
	fmt.Println("::: Introduce el nombre de la sala  :::")
	nombreSala, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	nombreSala = strings.TrimRight(nombreSala, "\n")
	nombreSala = strings.TrimSpace(nombreSala)
	conn.Write([]byte(nombreSala + "\n"))
}
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
func getHash(conn net.Conn) {
	fmt.Println("::: Enviando c. de seguridad :::")
	// Generate a new RSA key pair
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return
	}
	privateKey = privatekey
	// Publica
	publicKeySD := &privatekey.PublicKey
	pemKey := x509.MarshalPKCS1PublicKey(publicKeySD)
	keyss := keys{
		Publick: pemKey,
	}
	public, _ := json.Marshal(keyss)
	conn.Write(public)
	conn.Write([]byte("\n"))
}
func encriptar(msg []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	mensaje, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, msg, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return mensaje, nil
}
func desencriptar(msg []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	mensaje, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, msg, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return mensaje, nil
}
