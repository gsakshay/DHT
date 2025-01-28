package peer

import (
	"bufio"
	"dht/communication"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Peer represents an individual peer node in the DHT.
type Peer struct {
	ID               string
	Predecessor      string
	Successor        string
	StoreFilePath    string
	bootstrapAddress string
	communicator     *communication.TcpCommunicator
	mu               sync.Mutex
}

// NewPeer initializes a new peer with the given ID and communicator.
func NewPeer(id string, storeFilePath string, bootstrapAddress string, communicator *communication.TcpCommunicator) *Peer {
	return &Peer{
		ID:               id,
		StoreFilePath:    storeFilePath,
		communicator:     communicator,
		bootstrapAddress: bootstrapAddress,
	}
}

// JoinNetwork contacts the bootstrap server and registers the peer.
func (p *Peer) JoinNetwork(bootstrapAddress string) {
	// Send JOIN message to bootstrap server
	byteMessage, err := communication.GetJoinMessage(p.ID)
	if err == nil {
		p.communicator.SendMessage(bootstrapAddress, byteMessage)
	} else {
		fmt.Println("Error encoding join message:", err)
	}
}

// UpdateLinks updates the predecessor and successor of the peer.
func (p *Peer) UpdateLinks(predecessor, successor string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Predecessor = predecessor
	p.Successor = successor
	// Print once the links are updated
	fmt.Printf("Predecessor: %s, Successor: %s\n", p.Predecessor, p.Successor)
}
func (p *Peer) GetNeighbors() (string, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Predecessor, p.Successor
}

// StoreObject saves an object in the peer's local store.
func (p *Peer) StoreObject(reqID, clientID, objectID int) {
	// Get the current peerID and trim the first character from it and convert it to int
	nodeId, _ := strconv.Atoi(p.ID[1:])
	if objectID <= nodeId {
		// store it here
		// Open the file in append mode, create if not exists
		file, err := os.OpenFile(p.StoreFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening store file:", err)
			return
		}
		defer file.Close()

		// Write the object ID to the file
		entry := fmt.Sprintf("%d::%d\n", clientID, objectID)
		if _, err := file.WriteString(entry); err != nil {
			fmt.Println("Error writing to store file:", err)
		} else {
			// Send OBJ_STORED message to the bootstrap server
			byteMessage, err := communication.GetObjectStoredMessage(p.ID, objectID, clientID)
			if err == nil {
				go p.communicator.SendMessage(p.bootstrapAddress, byteMessage)
				// Print all the objects stored in the file
				// Open the file in read mode
				file, err := os.Open(p.StoreFilePath)
				if err != nil {
					fmt.Println("Error opening store file:", err)
					return
				}
				defer file.Close()

				// Read file line by line and check for the object
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					fmt.Println(line)
				}
			} else {
				fmt.Println("Error encoding obj stored message:", err)
			}
		}
	} else {
		// else forward it to the next peer
		go p.ForwardRequest(reqID, clientID, objectID, communication.STORE)
	}
}

// RetrieveObject fetches an object from the peer's store.
func (p *Peer) RetrieveObject(reqID, clientID, objectID int) {
	// Get the current peerID and trim the first character from it and convert it to int
	nodeId, _ := strconv.Atoi(p.ID[1:])
	if objectID <= nodeId {
		// Try retrieving the object from the local store

		// Open the file in read mode
		file, err := os.Open(p.StoreFilePath)
		if err != nil {
			fmt.Println("Error opening store file:", err)
			return
		}
		defer file.Close()

		// Read file line by line and check for the object
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.Split(line, "::")
			if len(parts) == 2 {
				savedClientID, _ := strconv.Atoi(parts[0])
				savedObjectID, _ := strconv.Atoi(parts[1])
				if savedClientID == clientID && savedObjectID == objectID {
					// Send OBJ_RETRIEVED message to the bootstrap server with status 1
					byteMessage, err := communication.GetObjectRetrievedMessage(1, objectID)
					if err == nil {
						go p.communicator.SendMessage(p.bootstrapAddress, byteMessage)
						return
					} else {
						fmt.Println("Error encoding obj retrieved message:", err)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading store file:", err)
		}

		// Send OBJ_RETRIEVED message to the bootstrap server with status -1
		byteMessage, err := communication.GetObjectRetrievedMessage(-1, objectID)
		if err == nil {
			go p.communicator.SendMessage(p.bootstrapAddress, byteMessage)
		}

	} else {
		// else forward it to the next peer
		go p.ForwardRequest(reqID, clientID, objectID, communication.RETRIEVE)
	}
}

// ForwardRequest forwards a lookup/store request to the appropriate peer in the ring.
func (p *Peer) ForwardRequest(reqID, clientID, objectID int, operationType communication.OperationType) {
	// Get the successor of the peer
	successor := p.Successor
	if successor == "" {
		fmt.Println("No successor found")
		return
	}

	// Send the request to the successor
	requestMessage, err := communication.GetRequestMessage(reqID, operationType, objectID, clientID)
	if err == nil {
		p.communicator.SendMessage(successor, requestMessage)
	} else {
		fmt.Println("Error encoding request message:", err)
	}
}
