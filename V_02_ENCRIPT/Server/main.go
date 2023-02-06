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

// This structure will contain the name of the room, the clients, the clients' public keys, the private key and the room's public key.
type room struct {
	nombre     string
	clientes   []net.Conn
	keys       []*rsa.PublicKey
	privateKey *rsa.PrivateKey
	publickKey *rsa.PublicKey
}

// We use this structure to receive the keys.
type keys struct {
	Publick []byte
}

// We use this structure to receive encrypted messages.
type msges struct {
	Mensaje []byte
}

// Array to store chat rooms.

var rooms []room

func main() {
	fmt.Println("Listening on the port: 8080")  // Descriptive message.
	conexion, err := net.Listen("tcp", ":8080") // We start listening over TCP on port 8080.
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conexion.Close()
	conns := make(map[net.Conn]string) //It serves to store connection data so that it can be used later.
	// If the connection is accepted, it will go to the function to manage connections, if there is an error, it will indicate it.
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
	hash, _ := bufio.NewReader(conexion).ReadString('\n') // Public key reader
	var Clavese keys
	err := json.Unmarshal([]byte(hash), &Clavese) // Parse to read the json
	if err != nil {
		fmt.Println(err)
		return
	}
	pubkey, _ := x509.ParsePKCS1PublicKey(Clavese.Publick)  // Get the connection's public key to use later
	opcion, _ := bufio.NewReader(conexion).ReadString('\n') // Option reader
	conexiones[conexion] = opcion
	fmt.Print("Te selected options is : " + opcion)
	nombreSala, _ := bufio.NewReader(conexion).ReadString('\n') // Room name reader
	conexiones[conexion] = nombreSala
	fmt.Print("Name of the room: " + nombreSala)
	joinRoom(conexion, nombreSala, pubkey)
	username, _ := bufio.NewReader(conexion).ReadString('\n') // User reader
	conexiones[conexion] = username
	fmt.Print("The user has entered to the room:" + username)
	for {
		mensajes, err := bufio.NewReader(conexion).ReadString('\n') // Message reader
		var message msges                                           // Initialize a variable that will be the json container
		err2 := json.Unmarshal([]byte(mensajes), &message)          // Parse the json to be able to read it
		if err2 != nil {
			fmt.Println(err)
			return
		}
		// Know the room's private key
		var r *room // Pointer to the room structure
		for i := range rooms {
			if rooms[i].nombre == nombreSala {
				r = &rooms[i] // Reference to the room we are using
				break
			}
		}
		fmt.Print("Room name:" + r.nombre)
		abc := desencriptar(message.Mensaje, r.privateKey) // Decrypt with the room's private key
		if err != nil {
			fmt.Println("Connection closed")
			return
		}
		user := strings.TrimSpace(conexiones[conexion])  // Clear whitespaces from the user
		msg := string(user + " : " + string(abc) + "\n") // Format text
		fmt.Print(msg)                                   // Display on screen
		broadcast(msg, nombreSala)                       // Send to all in the room
	}
}

// This function will try to create the room in the case it doesn't exist and if it does, it will add the user.
func joinRoom(conexion net.Conn, salaName string, pubkiclkey *rsa.PublicKey) {
	// We pass a parameter of the room name
	var r *room            // Pointer to the room structure
	for i := range rooms { // Rooms loop
		if rooms[i].nombre == salaName { // Searches for the room with the name passed by the function
			fmt.Println("The user has added to the room - " + salaName)
			r = &rooms[i]                                        // A reference is made to the room
			pemKey, _ := x509.MarshalPKIXPublicKey(r.publickKey) // The public key is parsed
			keyss := keys{                                       // Json object
				Publick: pemKey,
			}
			public, _ := json.Marshal(keyss) // It is transformed into a json object and parsed
			conexion.Write(public)           // It is sent to the client to obtain the server's public key
			conexion.Write([]byte("\n"))
			break
		}
	}
	// If it doesn't find the room, it will create it
	if r == nil {
		fmt.Println("The room has been created: " + salaName)
		p1, p2 := getHash(conexion)                                 // We get the room's private key and public key
		r = &room{nombre: salaName, privateKey: p1, publickKey: p2} // We create the room with the parameters we have obtained
		rooms = append(rooms, *r)                                   // We add it to the rooms array
		joinRoom(conexion, salaName, pubkiclkey)                    // We use this function for the client to be able to join the room after creating it
	}
	r.clientes = append(r.clientes, conexion) // We add the connections to the room we are in
	r.keys = append(r.keys, pubkiclkey)       // We add the public keys to the room we are in
}

// Sends a message to all users in the room passed as a parameter
func broadcast(mensaje string, nombreSala string) {
	for i := range rooms { // Searches in the rooms array
		if rooms[i].nombre == nombreSala { // Checks in which room we are
			for j := range rooms[i].clientes { // Gets each client
				msgEncript := encriptar([]byte(mensaje), rooms[i].keys[j])
				message := msges{
					Mensaje: msgEncript,
				}
				msgJson, _ := json.Marshal(message)
				rooms[i].clientes[j].Write(msgJson) // Sends the message for each client
				rooms[i].clientes[j].Write([]byte("\n"))
			}
			break
		}
	}
}

// Encrypts a message passed by byte with the destination's public
func encriptar(msg []byte, publicKey *rsa.PublicKey) []byte {
	mensaje, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, msg, nil) // Creates the encryption
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}
	return mensaje
}

// Decrypts a message passed by byte with the server's private
func desencriptar(msg []byte, privateKey *rsa.PrivateKey) []byte {
	mensaje, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, msg, nil) // Decrypts
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}
	return mensaje // Returns the decrypted message
}
func getHash(conn net.Conn) (*rsa.PrivateKey, *rsa.PublicKey) {
	fmt.Println("::: Sending security key :::")
	// Generate a new RSA key pair.
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
