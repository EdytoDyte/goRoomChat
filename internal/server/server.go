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

	// Read public key from client
	hash, _ := bufio.NewReader(conn).ReadString('\n')
	var keys chat.Keys
	if err := json.Unmarshal([]byte(hash), &keys); err != nil {
		fmt.Println(err)
		return
	}
	pubKey, _ := x509.ParsePKCS1PublicKey(keys.Publick)

	// Read room name from client
	roomName, _ := bufio.NewReader(conn).ReadString('\n')
	roomName = strings.TrimSpace(roomName)

	// Get or create room
	room := s.getOrCreateRoom(roomName, conn)

	// Read username from client
	username, _ := bufio.NewReader(conn).ReadString('\n')
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
		message, err := bufio.NewReader(conn).ReadString('\n')
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

		s.broadcast(fmt.Sprintf("%s: %s\n", username, string(decryptedMsg)), room)
	}
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
