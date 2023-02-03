package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	conns := make(map[net.Conn]string)
	go handleIncomingMessages(conn, conns)
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
	for {
		mensajes, err := bufio.NewReader(conexion).ReadString('\n')
		if err != nil {
			fmt.Println("Conexión cerrada")
			return
		}
		fmt.Print(mensajes)
	}
}
func getUser(conn net.Conn) {
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
	case "2":
		conn.Write([]byte("Unirse a una sala\n"))
		nombreSala(conn)
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

		conn.Write([]byte(mensajes + "\n"))
	}
}
