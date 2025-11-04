package tcp

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"

	"goflix/pkg/types"
)

// WriteMessage envía un mensaje con framing (4 bytes + JSON)
func WriteMessage(conn net.Conn, msg types.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))

	if _, err := conn.Write(length); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

func ReadMessage(conn net.Conn) (types.Message, error) {
	var msg types.Message

	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return msg, err
	}
	length := binary.BigEndian.Uint32(lenBuf)

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return msg, err
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}
