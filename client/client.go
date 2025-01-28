package client

import (
	"dht/communication"
	"fmt"
	"sync"
)

// Client represents a client interacting with the DHT
type Client struct {
	ID               int
	reqID            int
	bootstrapAddress string
	communicator     *communication.TcpCommunicator // Communicator for messaging
	mu               sync.Mutex
}

func NewClient(id int, bootstrapAddress string, communicator *communication.TcpCommunicator) *Client {
	return &Client{
		ID:               id,
		reqID:            1,
		bootstrapAddress: bootstrapAddress,
		communicator:     communicator,
	}
}

func (c *Client) RequestStore(objectID int) {
	c.mu.Lock()
	reqID := c.reqID
	c.reqID++ // Monotonically increasing
	c.mu.Unlock()

	requestMessage, err := communication.GetRequestMessage(reqID, communication.STORE, objectID, c.ID)

	if err == nil {
		err := c.communicator.SendMessage(c.bootstrapAddress, requestMessage)
		if err != nil {
			fmt.Println("Error sending request message:", err)
		}
	} else {
		fmt.Println("Error encoding request message:", err)
	}
}

func (c *Client) RequestRetrieve(objectID int) {
	c.mu.Lock()
	reqID := c.reqID
	c.reqID++ // Monotonically increasing
	c.mu.Unlock()

	requestMessage, err := communication.GetRequestMessage(reqID, communication.RETRIEVE, objectID, c.ID)

	if err == nil {
		err := c.communicator.SendMessage(c.bootstrapAddress, requestMessage)
		if err != nil {
			fmt.Println("Error sending request message:", err)
		}
	} else {
		fmt.Println("Error encoding request message:", err)
	}
}
