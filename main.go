package main

import (
	"dht/bootstrap"
	"dht/client"
	"dht/communication"
	"dht/peer"
	"dht/util"
	"fmt"
	"os"
	"time"
)

func main() {
	bootstrapName, objectFile, delay, testcase := util.ParseFlags()
	me, _ := os.Hostname()

	// Wait and block for the initial delay before proceeding with anything
	time.Sleep(time.Duration(delay) * time.Second)

	communicator := communication.NewTcpCommunicator(me)
	incomingMessagesCh := make(chan communication.Message)
	go communicator.Listen(incomingMessagesCh)

	var bootstrapObject *bootstrap.Bootstrap
	var clientObject *client.Client
	var peerObject *peer.Peer

	if me == "bootstrap" {
		bootstrapObject = bootstrap.NewBootstrap(communicator)
	} else if me == "client" {
		clientObject = client.NewClient(testcase-2, bootstrapName, communicator)
		if testcase == 3 {
			go clientObject.RequestStore(65) // 65 being the objectID
		} else if testcase == 4 {
			go clientObject.RequestRetrieve(66)
		} else if testcase == 5 {
			go clientObject.RequestRetrieve(110) // 110 not being in the ring
		}
	} else {
		peerObject = peer.NewPeer(me, objectFile, bootstrapName, communicator)
		peerObject.JoinNetwork(bootstrapName)
	}

	for message := range incomingMessagesCh {
		switch message.Header.Type {
		case communication.JOIN:
			if payload, ok := message.Payload.(*communication.JoinMessage); ok {
				go bootstrapObject.RegisterPeer(payload.PeerID)
			}
		case communication.RING:
			if payload, ok := message.Payload.(*communication.RingInformation); ok {
				peerObject.UpdateLinks(payload.Predecessor, payload.Successor)
			}
		case communication.REQUEST:
			if payload, ok := message.Payload.(*communication.RequestMessage); ok {
				if me == "bootstrap" {
					// Forward the request to the initial peer
					requestMessage, err := communication.GetRequestMessage(payload.ReqID, payload.OperationType, payload.ObjectID, payload.ClientID)
					if err != nil {
						fmt.Println("Error encoding request message:", err)
					} else {
						go communicator.SendMessage(bootstrapObject.GetFirstPeer(), requestMessage)
					}
				} else {
					// check the operation type and perform the operation
					if payload.OperationType == communication.STORE {
						peerObject.StoreObject(payload.ReqID, payload.ClientID, payload.ObjectID)
					} else if payload.OperationType == communication.RETRIEVE {
						peerObject.RetrieveObject(payload.ReqID, payload.ClientID, payload.ObjectID)
					} else {
						fmt.Println("Invalid operation type")
					}
				}
			}
		case communication.OBJ_STORED:
			if payload, ok := message.Payload.(*communication.ObjectStoredMessage); ok {
				if me == "bootstrap" {
					// Send the response back to the client
					responseMessage, err := communication.GetObjectStoredMessage(payload.PeerId, payload.ObjectID, payload.ClientID)
					if err != nil {
						fmt.Println("Error encoding response message:", err)
					} else {
						go communicator.SendMessage("client", responseMessage)
					}
				} else {
					// print the message
					fmt.Println("STORED: ", payload.ObjectID)
				}
			}
		case communication.OBJ_RETRIEVED:
			if payload, ok := message.Payload.(*communication.ObjectRetrievedMessage); ok {
				if me == "bootstrap" {
					// Send the response back to the client
					responseMessage, err := communication.GetObjectRetrievedMessage(payload.Status, payload.ObjectId)
					if err != nil {
						fmt.Println("Error encoding response message:", err)
					} else {
						go communicator.SendMessage("client", responseMessage)
					}
				} else {
					if payload.Status == -1 {
						fmt.Println("NOT FOUND: ", payload.ObjectId)
					} else {
						fmt.Println("RETRIEVED: ", payload.ObjectId)
					}
				}
			}
		}
	}
	// // Bootstrap
	// - The first one to start and Talks to both Peer and Client
	// - The first peer to join becomes point of contact for further actions
	// - Needs to maintain the correct ring structure
	// - When a peer contacts, it sends predecessor and successor and sends this peer as successor or predecessor to the existing one
	// - Stores the circular linkedList rightly
	// - Forwards client REQUEST to initial peer

	// // Peer
	// - Contacts the bootstrap server as soon as it starts and stores the predecessor and successor
	// - Can receive messages from bootstrap about updated predecessor and successor
	// - Peer should be able to forward the REQUEST to successor
	// - If a request comes to initial peer from someone apart from bootstrap, then the object does not exist in the ring and would send -1 to the bootstrap
	// - Peer should be able to STORE and RETRIEVE the object given ObjectId and ClientId
	// - Sends OBJ_STORED message back to the bootstrap

	// // Client
	// - Sends a REQUEST message to the bootstrap server
}
