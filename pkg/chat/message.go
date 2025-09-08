package chat

import (
	"crypto/rsa"
	"net"
)

// This structure will contain the name of the room, the clients, the clients' public keys, the private key and the room's public key.
type Room struct {
	Name       string
	Clients    map[net.Conn]*Client
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

type Client struct {
	Conn     net.Conn
	Name     string
	Room     *Room
	PubKey   *rsa.PublicKey
}

// We use this structure to receive the keys.
type Keys struct {
	Protocol []byte
	Publick  []byte
}

// We use this structure to receive encrypted messages.
type Msges struct {
	Protocol []byte
	Mensaje  []byte
}
