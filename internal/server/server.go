package server

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
	"sync"

	"github.com/Go-Chat/pkg/chat"
)

type Server struct {
	listener net.Listener
	rooms    map[string]*chat.Room
	mutex    sync.Mutex
}

func NewServer() (*Server, error) {
	return &Server{
		rooms: make(map[string]*chat.Room),
	}, nil
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	// Read public key from client
	hash, _ := reader.ReadString('\n')
	fmt.Println("Server: Received data from client (potential public key).")
	var keys chat.Keys
	if err := json.Unmarshal([]byte(hash), &keys); err != nil {
		fmt.Println(err)
		return
	}
	pubKey, _ := x509.ParsePKCS1PublicKey(keys.Publick)

	// Read room name from client
	roomName, _ := reader.ReadString('\n')
	roomName = strings.TrimSpace(roomName)
	fmt.Println("Server: Received room name:", roomName)
	
	// Get or create room
	room := s.getOrCreateRoom(roomName, conn)

	// Read username from client
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	client := &chat.Client{
		Conn:   conn,
		Name:   username,
		Room:   room,
		PubKey: pubKey,
	}
	room.Clients[conn] = client

	s.broadcast(fmt.Sprintf("::: %s has joined the room :::\n", username), room)

	for {
		// Read message from client
		message, err := reader.ReadString('\n')
		if err != nil {
			s.broadcast(fmt.Sprintf("::: %s has left the room :::\n", username), room)
			delete(room.Clients, conn)
			return
		}

		var msg chat.Msges
		if err := json.Unmarshal([]byte(message), &msg); err != nil {
			fmt.Println(err)
			continue
		}

		decryptedMsg, err := s.decrypt(msg.Mensaje, room.PrivateKey)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if strings.HasPrefix(string(decryptedMsg), "/rooms") {
			s.listRooms(client)
			continue
		}

		if strings.HasPrefix(string(decryptedMsg), "/users") {
			s.listUsers(client)
			continue
		}

		if strings.HasPrefix(string(decryptedMsg), "/msg") {
			parts := strings.SplitN(string(decryptedMsg), " ", 3)
			if len(parts) < 3 {
				s.sendMessage("Usage: /msg <user> <message>", client)
				continue
			}
			s.privateMessage(parts[1], parts[2], client)
			continue
		}

		s.broadcast(fmt.Sprintf("%s: %s\n", username, string(decryptedMsg)), room)
	}
}

func (s *Server) listRooms(client *chat.Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	msg := strings.Join(rooms, ", ")
	encryptedMsg, err := s.encrypt([]byte(msg), client.PubKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	msgJSON, _ := json.Marshal(chat.Msges{
		Protocol: []byte("msg"),
		Mensaje:  encryptedMsg,
	})
	client.Conn.Write(msgJSON)
	client.Conn.Write([]byte("\n"))
}

func (s *Server) privateMessage(username, message string, sender *chat.Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, client := range sender.Room.Clients {
		if client.Name == username {
			s.sendMessage(fmt.Sprintf("(private) %s: %s", sender.Name, message), client)
			return
		}
	}
	s.sendMessage(fmt.Sprintf("User %s not found", username), sender)
}

func (s *Server) sendMessage(message string, client *chat.Client) {
	encryptedMsg, err := s.encrypt([]byte(message), client.PubKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	msgJSON, _ := json.Marshal(chat.Msges{
		Protocol: []byte("msg"),
		Mensaje:  encryptedMsg,
	})
	client.Conn.Write(msgJSON)
	client.Conn.Write([]byte("\n"))
}

func (s *Server) listUsers(client *chat.Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var users []string
	for _, c := range client.Room.Clients {
		users = append(users, c.Name)
	}

	msg := strings.Join(users, ", ")
	encryptedMsg, err := s.encrypt([]byte(msg), client.PubKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	msgJSON, _ := json.Marshal(chat.Msges{
		Protocol: []byte("msg"),
		Mensaje:  encryptedMsg,
	})
	client.Conn.Write(msgJSON)
	client.Conn.Write([]byte("\n"))
}

func (s *Server) getOrCreateRoom(roomName string, conn net.Conn) *chat.Room {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, ok := s.rooms[roomName]
	if !ok {
		privateKey, publicKey := s.generateKeys(conn)
		room = &chat.Room{
			Name:       roomName,
			Clients:    make(map[net.Conn]*chat.Client),
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		}
		s.rooms[roomName] = room
	}

	// Send room public key to client
	fmt.Println("Server: Sending room key to client.")
	pemKey, _ := x509.MarshalPKIXPublicKey(room.PublicKey)
	keyss := chat.Keys{
		Protocol: []byte("key"),
		Publick:  pemKey,
	}
	public, _ := json.Marshal(keyss)
	conn.Write(public)
	conn.Write([]byte("\n"))

	return room
}

func (s *Server) broadcast(message string, room *chat.Room) {
	for _, client := range room.Clients {
		encryptedMsg, err := s.encrypt([]byte(message), client.PubKey)
		if err != nil {
			fmt.Println(err)
			continue
		}
		msg := chat.Msges{
			Protocol: []byte("msg"),
			Mensaje:  encryptedMsg,
		}
		msgJson, _ := json.Marshal(msg)
		client.Conn.Write(msgJson)
		client.Conn.Write([]byte("\n"))
	}
}

func (s *Server) encrypt(msg []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, msg, nil)
}

func (s *Server) decrypt(msg []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, msg, nil)
}

func (s *Server) generateKeys(conn net.Conn) (*rsa.PrivateKey, *rsa.PublicKey) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	return privatekey, &privatekey.PublicKey
}
