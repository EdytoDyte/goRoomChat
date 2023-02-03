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
	"strings"
)

// Esta estructura contendra el nombre de la sala,los clientes, las publick keys de los clientes, la privatekey y la publickkey de la sala
type room struct {
	nombre     string
	clientes   []net.Conn
	keys       []*rsa.PublicKey
	privateKey *rsa.PrivateKey
	publickKey *rsa.PublicKey
}

// Utilizamos esta estructura para poder recibir las claves
type keys struct {
	Publick []byte
}

// Utilizamos esta estructura para poder recibir mensajes encriptados
type msges struct {
	Mensaje []byte
}

// Array para poder almacenar las salas de chat
var rooms []room

func main() {
	fmt.Println("Escuchando desde el servidor en el puerto 8080") // Mensaje descriptivo
	conexion, err := net.Listen("tcp", ":8080")                   // Empezamos a esuchar por tcp en el peurto 8080
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conexion.Close()
	conns := make(map[net.Conn]string) // Sirve para almacenar datos de la conexion para poder usarlos despues
	// Si se acepta la conexion se ira a la funcion para gestionar las conexiones si hay un error lo indicara
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
	hash, _ := bufio.NewReader(conexion).ReadString('\n') // Lector de claves publicas
	var Clavese keys
	err := json.Unmarshal([]byte(hash), &Clavese) // Parsear para poder leer el json
	if err != nil {
		fmt.Println(err)
		return
	}
	pubkey, _ := x509.ParsePKCS1PublicKey(Clavese.Publick)  // Obtener la publickey de la conexion para poder usarla despues
	opcion, _ := bufio.NewReader(conexion).ReadString('\n') // Lector de opcion
	conexiones[conexion] = opcion
	fmt.Print("Se ha selecionado la opcion: " + opcion)
	nombreSala, _ := bufio.NewReader(conexion).ReadString('\n') // Lector del nombre de sala
	conexiones[conexion] = nombreSala
	fmt.Print("Nombre de sala: " + nombreSala)
	joinRoom(conexion, nombreSala, pubkey)
	username, _ := bufio.NewReader(conexion).ReadString('\n') // Lector de usuarios
	conexiones[conexion] = username
	fmt.Print("Se ha unido al chat  " + username)
	for {
		mensajes, err := bufio.NewReader(conexion).ReadString('\n') // Lector de mensajes
		var message msges                                           // Inicializar una variable que sera el contenedor de json
		err2 := json.Unmarshal([]byte(mensajes), &message)          // Parsear el json para poder leerlos
		if err2 != nil {
			fmt.Println(err)
			return
		}
		// Saber la privateKey de la room
		var r *room // puntero a la estructura room
		for i := range rooms {
			if rooms[i].nombre == nombreSala {
				r = &rooms[i] // Referencia a la sala que vamos a usar
				break
			}
		}
		abc := desencriptar(message.Mensaje, r.privateKey) // Desencriptar con la privatekey de la sala
		fmt.Print(string(abc) + "\n")
		if err != nil {
			fmt.Println("Conexión cerrada")
			return
		}
		user := strings.TrimSpace(conexiones[conexion])  // Limpiar espacios en blanco del usuario
		msg := string(user + " : " + string(abc) + "\n") // Formatear texto
		fmt.Print(msg)                                   // Mostrar por pantalla
		broadcast(msg, nombreSala)                       // Enviar a todos los de la sala
	}
}

// Esta funcion tratara de crear la sala en el caso de que no exista y en el que si pues añadira al usuario
func joinRoom(conexion net.Conn, salaName string, pubkiclkey *rsa.PublicKey) {
	// Le pasamos un parametro del nombre de la sala
	var r *room            // puntero a la estructura room
	for i := range rooms { // Bucle de las salas
		if rooms[i].nombre == salaName { // Busca la sala que tiene el nombre pasado por la funcioon
			fmt.Println("Se ha añadido a la sala " + salaName)
			r = &rooms[i]                                        // Se hace un referencia a la sala
			pemKey, _ := x509.MarshalPKIXPublicKey(r.publickKey) // La clave publicka se parsea
			keyss := keys{                                       // Objeto Json
				Publick: pemKey,
			}
			public, _ := json.Marshal(keyss) // Se transforma en objeto json y se parsea a el
			conexion.Write(public)           // Se le manda al cliente para poder obtener la clave publica del servidor
			conexion.Write([]byte("\n"))
			break
		}
	}
	// Si no encuentra la sala, la creara
	if r == nil {
		fmt.Println("Se ha creado la sala " + salaName)
		p1, p2 := getHash(conexion)                                 // Obtenemos la private key y la publick key de la sala
		r = &room{nombre: salaName, privateKey: p1, publickKey: p2} // Creamos la sala con los parametros que hemos obtenido
		rooms = append(rooms, *r)                                   // La añadimos a la array de salas
		joinRoom(conexion, salaName, pubkiclkey)                    // Utilizamos esta funcion para que el cliente se pueda unir a la sala despues de crearla
	}
	r.clientes = append(r.clientes, conexion) // Añadimos las conexiones a la sala en que estamos
	r.keys = append(r.keys, pubkiclkey)       // Añadimos las publick keys a la sala en que estamos
}

// Envia un mensaje a todos los usuarios de la sala que se ha pasado como parametro
func broadcast(mensaje string, nombreSala string) {
	for i := range rooms { // busca en la array de salas
		if rooms[i].nombre == nombreSala { // Comprueba en que sala estamos
			for j := range rooms[i].clientes { // Se obtiene cada cliente
				msgEncript := encriptar([]byte(mensaje), rooms[i].keys[j])
				message := msges{
					Mensaje: msgEncript,
				}
				msgJson, _ := json.Marshal(message)
				rooms[i].clientes[j].Write(msgJson)      // Se envia el mensaje para cada cliente
				rooms[i].clientes[j].Write([]byte("\n")) // Se envia el mensaje para cada cliente

			}
			break
		}
	}
}

// Encripta un mensaje que haya sido pasado por byte con la publica del destino
func encriptar(msg []byte, publicKey *rsa.PublicKey) []byte {
	mensaje, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, msg, nil) // Crea la encriptacion
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}
	return mensaje
}

// Desencripta un mensaje que haya sido pasado por byte con la privada del servidor
func desencriptar(msg []byte, privateKey *rsa.PrivateKey) []byte {
	mensaje, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, msg, nil) // Desencripta
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}
	return mensaje // Devuelve el mensaje desencriptado
}
func getHash(conn net.Conn) (*rsa.PrivateKey, *rsa.PublicKey) {
	fmt.Println("::: Enviando c. de seguridad :::")
	// Generate a new RSA key pair
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	publickey := &privatekey.PublicKey // <-- clave publica
	pemKey, _ := x509.MarshalPKIXPublicKey(publickey)
	keyss := keys{
		Publick: pemKey,
	}
	public, _ := json.Marshal(keyss)
	conn.Write(public)
	conn.Write([]byte("\n"))
	return privatekey, publickey

}
