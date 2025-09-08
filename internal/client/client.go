package client

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

	"github.com/Go-Chat/pkg/chat"
)

type Client struct {
	conn       net.Conn
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	serverKey  *rsa.PublicKey
	ui         *UI
}

func NewClient() (*Client, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &Client{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

func (c *Client) Start(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.conn = conn
	defer c.conn.Close()

	go c.handleIncomingMessages()

	c.ui = NewUI(c)
	if err := c.ui.Run(); err != nil {
		return err
	}

	return nil
}

func (c *Client) sendPublicKey() {
	pemKey := x509.MarshalPKCS1PublicKey(c.publicKey)
	keyss := chat.Keys{
		Publick: pemKey,
	}
	public, _ := json.Marshal(keyss)
	c.conn.Write(public)
	c.conn.Write([]byte("\n"))
}

func (c *Client) handleIncomingMessages() {
	reader := bufio.NewReader(c.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed")
			os.Exit(0)
		}
		var proto struct{ Protocol []byte }
		if err := json.Unmarshal([]byte(message), &proto); err != nil {
			continue
		}

		switch string(proto.Protocol) {
		case "key":
			var keys chat.Keys
			if err := json.Unmarshal([]byte(message), &keys); err == nil {
				pubkey, _ := x509.ParsePKIXPublicKey(keys.Publick)
				c.serverKey = pubkey.(*rsa.PublicKey)
				c.ui.SetServerKey(true)
			}
		case "msg":
			var msg chat.Msges
			if err := json.Unmarshal([]byte(message), &msg); err == nil {
				decryptedMsg, err := c.decrypt(msg.Mensaje)
				if err != nil {
					fmt.Println(err)
					continue
				}
				c.ui.UpdateMessages(string(decryptedMsg))
			}
		}
	}
}

func (c *Client) SendMessage(message string) {
	encryptedMsg, err := c.encrypt([]byte(message))
	if err != nil {
		fmt.Println(err)
		return
	}
	msg := chat.Msges{
		Protocol: []byte("msg"),
		Mensaje:  encryptedMsg,
	}
	msgJson, _ := json.Marshal(msg)
	c.conn.Write(msgJson)
	c.conn.Write([]byte("\n"))
}

func (c *Client) JoinRoom(roomName string) {
	pemKey := x509.MarshalPKCS1PublicKey(c.publicKey)
	keyss := chat.Keys{
		Publick: pemKey,
	}
	public, _ := json.Marshal(keyss)
	c.conn.Write(public)
	c.conn.Write([]byte("\n"))
	c.conn.Write([]byte(roomName + "\n"))
}

func (c *Client) SendUsername(username string) {
	c.conn.Write([]byte(username + "\n"))
}

func (c *Client) encrypt(msg []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, c.serverKey, msg, nil)
}

func (c *Client) decrypt(msg []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, c.privateKey, msg, nil)
}
