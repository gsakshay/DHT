package bootstrap

import (
	"dht/communication"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Bootstrap maintains the peer ring
type Bootstrap struct {
	peers        []string                       // Sorted slice of peer IDs
	mu           sync.Mutex                     // Mutex for thread safety
	communicator *communication.TcpCommunicator // Communicator for messaging peers
}

// NewBootstrap initializes the bootstrap server
func NewBootstrap(communicator *communication.TcpCommunicator) *Bootstrap {
	return &Bootstrap{
		peers:        []string{},
		communicator: communicator,
	}
}

func (b *Bootstrap) GetFirstPeer() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.peers[0]
}

// extractNumber removes the 'n' prefix and converts the rest to an integer
func extractNumber(peerID string) int {
	numStr := strings.TrimPrefix(peerID, "n")
	num, _ := strconv.Atoi(numStr)
	return num
}

// RegisterPeer adds a new peer while keeping the list numerically sorted
func (b *Bootstrap) RegisterPeer(peerID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Insert the new peer
	b.peers = append(b.peers, peerID)

	// Sort using numeric order by ignoring "n" prefix
	sort.Slice(b.peers, func(i, j int) bool {
		return extractNumber(b.peers[i]) < extractNumber(b.peers[j])
	})

	// Find index of the new peer
	index := sort.Search(len(b.peers), func(i int) bool {
		return extractNumber(b.peers[i]) >= extractNumber(peerID)
	})

	// Determine predecessor and successor
	predecessor, successor := b.getNeighbors(index)

	// print the ring once
	fmt.Println("Ring: ", b.peers)

	// Notify affected peers
	go b.notifyNeighbors(peerID, predecessor, successor)
}

func (b *Bootstrap) getIndex(peerID string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Find index of the peer
	index := sort.Search(len(b.peers), func(i int) bool {
		return extractNumber(b.peers[i]) >= extractNumber(peerID)
	})

	return index
}

// getNeighbors finds the predecessor and successor for a given index
func (b *Bootstrap) getNeighbors(index int) (string, string) {
	n := len(b.peers)
	if n == 1 {
		return b.peers[0], b.peers[0] // Only one peer in the ring, points to itself
	}
	predecessor := b.peers[(index-1+n)%n]
	successor := b.peers[(index+1)%n]
	return predecessor, successor
}

// notifyNeighbors informs affected peers of changes
func (b *Bootstrap) notifyNeighbors(peerID, predecessor, successor string) {
	// Send message to peerID with predecessor and successor
	ringMessage, err := communication.GetRingMessage(predecessor, successor)
	if err == nil {
		err := b.communicator.SendMessage(peerID, ringMessage)
		if err != nil {
			fmt.Println("Error sending ring message:", err)
		}
	} else {
		fmt.Println("Error encoding join message:", err)
	}

	if predecessor != peerID {
		predecessorIndex := b.getIndex(predecessor)
		predecessorPredecessor, predecessorSuccessor := b.getNeighbors(predecessorIndex)
		ringMessagePredecessorUpdate, err := communication.GetRingMessage(predecessorPredecessor, predecessorSuccessor)
		if err == nil {
			err := b.communicator.SendMessage(predecessor, ringMessagePredecessorUpdate)
			if err != nil {
				fmt.Println("Error sending ring message:", err)
			}
		} else {
			fmt.Println("Error encoding join message:", err)
		}
	}

	if successor != peerID && successor != predecessor {
		// Send message to successor with new predecessor
		successorIndex := b.getIndex(successor)
		successorPredecessor, successorSuccessor := b.getNeighbors(successorIndex)
		ringMessagePredecessorUpdate, err := communication.GetRingMessage(successorPredecessor, successorSuccessor)
		if err == nil {
			err := b.communicator.SendMessage(successor, ringMessagePredecessorUpdate)
			if err != nil {
				fmt.Println("Error sending ring message:", err)
			}
		} else {
			fmt.Println("Error encoding join message:", err)
		}
	}
}
