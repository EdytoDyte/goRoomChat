package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type room struct {
	nombre   string
	clientes []net.Conn
}

var rooms []room

func main() {
	fmt.Println("Escuchando desde el servidor en el puerto 8080")
	conexion, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conexion.Close()
	conns := make(map[net.Conn]string)
	for {
		conn, err := conexion.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn, conns)
	}
}
func handleConnection(conexion net.Conn, conexiones map[net.Conn]string) {
	defer conexion.Close()
	opcion, _ := bufio.NewReader(conexion).ReadString('\n')
	conexiones[conexion] = opcion
	fmt.Print("Se ha selecionado la opcion: " + opcion)
	nombreSala, _ := bufio.NewReader(conexion).ReadString('\n')
	conexiones[conexion] = nombreSala
	fmt.Print("Nombre de sala: " + nombreSala)
	joinRoom(conexion, nombreSala)
	username, _ := bufio.NewReader(conexion).ReadString('\n')
	conexiones[conexion] = username
	fmt.Print("Se ha unido al chat  " + username)
	for {
		mensajes, err := bufio.NewReader(conexion).ReadString('\n')
		mensajes = strings.TrimRight(mensajes, "\n")
		mensajes = strings.TrimSpace(mensajes)
		if err != nil {
			fmt.Println("Conexión cerrada")
			return
		}
		user := strings.TrimSpace(conexiones[conexion])
		msg := string(user + " : " + mensajes + "\n")
		fmt.Print(msg)
		broadcast(msg, nombreSala)
	}

}
func joinRoom(conexion net.Conn, salaName string) {
	// Le pasamos un parametro del nombre de la sala
	var r *room // puntero a la estructura room
	for i := range rooms {
		if rooms[i].nombre == salaName {
			fmt.Println("Se ha añadido a la sala " + salaName)
			r = &rooms[i]
			break
		}
	}
	// Si no encuentra la sala, la creara
	if r == nil {
		fmt.Println("Se ha creado la sala " + salaName)
		r = &room{nombre: salaName}
		rooms = append(rooms, *r)
		joinRoom(conexion, salaName)
	}
	r.clientes = append(r.clientes, conexion)
}
func broadcast(mensaje string, nombreSala string) {
	for i := range rooms {
		if rooms[i].nombre == nombreSala {
			for j := range rooms[i].clientes {
				rooms[i].clientes[j].Write([]byte(mensaje))
			}
			break
		}
	}
}
