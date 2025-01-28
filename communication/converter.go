package communication

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"net"
)

type MessageType uint8
type OperationType uint8

const (
	JOIN MessageType = iota
	RING
	REQUEST
	OBJ_STORED
	OBJ_RETRIEVED
)

const (
	STORE OperationType = iota
	RETRIEVE
)

type MessageHeader struct {
	Type   MessageType
	Length uint32
}

type Message struct {
	Header  MessageHeader
	Payload interface{}
}

type JoinMessage struct {
	PeerID string
}

type RingInformation struct {
	Predecessor string
	Successor   string
}

type RequestMessage struct {
	ReqID         int
	OperationType OperationType
	ObjectID      int
	ClientID      int
}

type ObjectStoredMessage struct {
	PeerId   string
	ObjectID int
	ClientID int
}

type ObjectRetrievedMessage struct {
	Status   int
	ObjectId interface{}
}

// Register types before decoding
func init() {
	gob.Register(JoinMessage{})
	gob.Register(RingInformation{})
	gob.Register(RequestMessage{})
	gob.Register(ObjectStoredMessage{})
	gob.Register(ObjectRetrievedMessage{})
}

func encodeMessage(msg Message) ([]byte, error) {
	// Encode payload to binary
	payloadBytes, err := EncodeToBinary(msg.Payload)
	if err != nil {
		return nil, err
	}

	// Prepare buffer
	buf := new(bytes.Buffer)

	// Write header
	if err := binary.Write(buf, binary.LittleEndian, msg.Header.Type); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(payloadBytes))); err != nil {
		return nil, err
	}

	// Write payload
	if _, err := buf.Write(payloadBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ReadMessage reads a complete message from the connection
func ReadMessage(conn net.Conn) (*Message, error) {
	// Read header first
	header := MessageHeader{}
	headerBytes := make([]byte, 5) // 1 byte type + 4 bytes length

	// Ensure we read the entire header
	_, err := io.ReadFull(conn, headerBytes)
	if err != nil {
		fmt.Println("Error reading header:", err)
		return nil, err
	}

	// Parse header
	header.Type = MessageType(headerBytes[0])
	header.Length = binary.LittleEndian.Uint32(headerBytes[1:5])

	// Read payload
	payloadBytes := make([]byte, header.Length)
	_, err = io.ReadFull(conn, payloadBytes)
	if err != nil {
		fmt.Println("Error reading payload:", err)
		return nil, err
	}

	// Decode payload (use appropriate type based on header)
	var payload interface{}
	switch header.Type {
	case JOIN:
		payload = &JoinMessage{}
	case RING:
		payload = &RingInformation{}
	case REQUEST:
		payload = &RequestMessage{}
	case OBJ_STORED:
		payload = &ObjectStoredMessage{}
	case OBJ_RETRIEVED:
		payload = &ObjectRetrievedMessage{}
	default:
		err := fmt.Errorf("unknown message type")
		fmt.Println(err)
		return nil, err
	}

	// Decode payload
	err = DecodeFromBinary(payloadBytes, payload)
	if err != nil {
		fmt.Println("Error decoding payload:", err)
		return nil, err
	}

	return &Message{
		Header:  header,
		Payload: payload,
	}, nil
}

// EncodeToBinary converts a struct to binary bytes using Gob
func EncodeToBinary(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(data)
	return buf.Bytes(), err
}

// DecodeFromBinary converts binary bytes to a struct using Gob
func DecodeFromBinary(b []byte, data interface{}) error {
	buf := bytes.NewReader(b)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(data)
}

func GetJoinMessage(peerID string) ([]byte, error) {
	joinMsg := JoinMessage{
		PeerID: peerID,
	}
	msg := Message{
		Header: MessageHeader{
			Type:   JOIN,
			Length: uint32(binary.Size(joinMsg)),
		},
		Payload: joinMsg,
	}
	byteMessage, err := encodeMessage(msg)
	if err != nil {
		return nil, err
	}

	return byteMessage, nil
}

func GetRingMessage(predecessor, successor string) ([]byte, error) {
	ringInfo := RingInformation{
		Predecessor: predecessor,
		Successor:   successor,
	}
	msg := Message{
		Header: MessageHeader{
			Type:   RING,
			Length: uint32(binary.Size(ringInfo)),
		},
		Payload: ringInfo,
	}
	byteMessage, err := encodeMessage(msg)
	if err != nil {
		return nil, err
	}

	return byteMessage, nil
}

func GetRequestMessage(reqID int, operationType OperationType, objectID int, clientID int) ([]byte, error) {
	reqMsg := RequestMessage{
		ReqID:         reqID,
		OperationType: operationType,
		ObjectID:      objectID,
		ClientID:      clientID,
	}
	msg := Message{
		Header: MessageHeader{
			Type:   REQUEST,
			Length: uint32(binary.Size(reqMsg)),
		},
		Payload: reqMsg,
	}
	byteMessage, err := encodeMessage(msg)
	if err != nil {
		return nil, err
	}

	return byteMessage, nil
}

func GetObjectStoredMessage(peerID string, objectID int, clientID int) ([]byte, error) {
	objStoredMsg := ObjectStoredMessage{
		PeerId:   peerID,
		ObjectID: objectID,
		ClientID: clientID,
	}
	msg := Message{
		Header: MessageHeader{
			Type:   OBJ_STORED,
			Length: uint32(binary.Size(objStoredMsg)),
		},
		Payload: objStoredMsg,
	}
	byteMessage, err := encodeMessage(msg)
	if err != nil {
		return nil, err
	}

	return byteMessage, nil
}

func GetObjectRetrievedMessage(status int, objectID interface{}) ([]byte, error) {
	objRetrievedMsg := ObjectRetrievedMessage{
		Status:   status,
		ObjectId: objectID,
	}
	msg := Message{
		Header: MessageHeader{
			Type:   OBJ_RETRIEVED,
			Length: uint32(binary.Size(objRetrievedMsg)),
		},
		Payload: objRetrievedMsg,
	}
	byteMessage, err := encodeMessage(msg)
	if err != nil {
		return nil, err
	}

	return byteMessage, nil
}
