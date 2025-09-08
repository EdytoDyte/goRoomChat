# Go Chat

Go Chat is a simple, secure, and real-time chat application written in Go. It allows users to create or join chat rooms and communicate with each other using encrypted messages.

## Features

- **Secure Communication:** All messages are end-to-end encrypted using RSA encryption.
- **Chat Rooms:** Create your own chat rooms or join existing ones.
- **Real-time Messaging:** Messages are delivered in real-time to all users in a room.
- **List Rooms:** Get a list of all available chat rooms.
- **List Users:** Get a list of all users in the current chat room.
- **Private Messaging:** Send private messages to other users in the same room.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) (version 1.15 or later)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/EdytoDyte/goRoomChat.git
   ```
2. Navigate to the project directory:
   ```bash
   cd goRoomChat
   ```

### Running the Application

1. **Start the server:**
   ```bash
   go run cmd/server/main.go
   ```
   The server will start listening on port `8080`.

2. **Start the client:**
   ```bash
   go run cmd/client/main.go
   ```
   A new terminal window will open with the chat client.

## How to Use

1. **Enter a room name:** When you start the client, you will be prompted to enter a room name. If the room doesn't exist, it will be created for you.

2. **Enter a username:** After entering a room name, you will be prompted to enter a username. This will be your display name in the chat room.

3. **Chat:** Once you've entered a room and username, you can start sending messages to the room.

### Commands

- `/rooms`: Get a list of all available chat rooms.
- `/users`: Get a list of all users in the current chat room.
- `/msg <user> <message>`: Send a private message to the specified user.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
