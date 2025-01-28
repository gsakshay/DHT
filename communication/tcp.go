package communication

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const TCPPort = "8888"

type TcpCommunicator struct {
	selfId      string              // The ID of the current peer.
	connections map[string]net.Conn // Maps peer IDs to their active TCP connections.
	mu          sync.Mutex          // Mutex for thread-safe access to connections.
}

func NewTcpCommunicator(id string) *TcpCommunicator {
	return &TcpCommunicator{
		selfId:      id,
		connections: make(map[string]net.Conn),
	}
}

// SendMessage dynamically establishes a connection if one does not exist and then sends the message.
func (c *TcpCommunicator) SendMessage(to string, message []byte) error {
	c.mu.Lock()
	conn, exists := c.connections[to]
	c.mu.Unlock()

	if !exists {
		// Establish the connection if it does not exist
		var err error
		conn, err = c.establishConnection(to)
		if err != nil {
			return fmt.Errorf("failed to establish connection to peer %s: %w", to, err)
		} else {
			// Store the connection for future use
			c.mu.Lock()
			c.connections[to] = conn
			c.mu.Unlock()
		}
	}

	// Send the message
	_, err := conn.Write(message)
	if err != nil {
		log.Printf("Failed to send message to peer %s: %v", to, err)
		// If sending fails, remove the connection to force re-establishment later
		c.mu.Lock()
		delete(c.connections, to)
		c.mu.Unlock()
		return fmt.Errorf("failed to send message to peer %s: %w", to, err)
	}

	return nil
}

// establishConnection establishes a TCP connection to a specific peer and stores it in the connections map.
func (c *TcpCommunicator) establishConnection(address string) (net.Conn, error) {
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", address, TCPPort))
		if err == nil {
			return conn, nil
		}

		// Retry after a short delay if the connection fails
		time.Sleep(500 * time.Millisecond)
	}
}

func (c *TcpCommunicator) Listen(messageCh chan Message) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", TCPPort))
	if err != nil {
		log.Fatalf("Failed to start listener on port %s: %v", TCPPort, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go func(conn net.Conn) {
			defer func() {
				conn.Close()
			}()
			for {
				// Read and parse the message
				fullMessage, err := ReadMessage(conn)
				if err != nil {
					log.Printf("Failed to read and parse message from %s: %v", conn.RemoteAddr().String(), err)
					break
				}
				// Handle the message
				messageCh <- *fullMessage
			}
		}(conn)
	}
}
