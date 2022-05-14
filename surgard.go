package surgard

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

const cmdLayout = "5%02d%01d 18%04d%01s%03d%02d%03d%s"
const eoc = "\x14"

//
// client
//

type Client struct {
	addr string
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return &Client{addr: addr}, nil
}

func (c *Client) Send(cmd string) error {
	conn, err := net.DialTimeout("tcp", c.addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(cmd + eoc)); err != nil {
		return fmt.Errorf("failed to send re-command, %v", err)
	}

	// read response, should be 1 byte, 0x06 - command recieve, 0x15 - need to repeat
	b := make([]byte, 1)
	if _, err := conn.Read(b); err != nil {
		return fmt.Errorf("failed to read response, %v", err)
	}

	// repeat command if it required
	if b[0] == 0x15 {
		if _, err := conn.Write([]byte(cmd + eoc)); err != nil {
			return fmt.Errorf("failed to send re-command, %v", err)
		}
	}

	// success
	if b[0] == 0x06 {
		return nil
	}

	return nil
}

func (c *Client) Dial(d DialData) error {
	if err := d.Validate(); err != nil {
		return err
	}

	q := "E"
	if d.Close {
		q = "R"
	}

	cmd := fmt.Sprintf(cmdLayout, d.PP, d.R, d.Object, q, d.Code, d.Group, d.Zone, d.Tail)

	return c.Send(cmd)
}

//
// server
//

type Receiver struct {
	addr    string
	handler ReceiverHandler
}

type ReceiverHandler func(data DialData)

func NewReceiver(addr string, handler func(data DialData)) (r *Receiver, err error) {
	r = &Receiver{addr, handler}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return r, err
	}

	log.Println("new receiver:", addr)

	for {
		conn, err := ln.Accept()
		log.Println("connection", conn.RemoteAddr())

		if err != nil {
			log.Println(err)
			continue
		}

		go r.handleConn(conn)
	}
}

var confirm = []byte{0x06}

func (r *Receiver) handleConn(conn net.Conn) {
	for {
		b := make([]byte, 128)

		n, err := conn.Read(b)
		if err != nil {
			break
		}

		// ping
		if b[0] == 0x31 {
			conn.Write(confirm)
			continue
		}

		var d DialData
		var close string
		if n, err := fmt.Sscanf(string(b[:n]), cmdLayout, &d.PP, &d.R, &d.Object, &close, &d.Code, &d.Group, &d.Zone, &d.Tail); err != nil {
			log.Println("failed to parse incoming message:")
			log.Printf("%v, %d, %+v", err, n, d)
			fmt.Println(hex.Dump(b[:n]))
			continue
		}

		if close == "R" {
			d.Close = true
		}

		r.handler(d)

		// confirm message
		conn.Write(confirm)
	}
}

//
// data
//

type DialData struct {
	PP uint
	R  uint

	Object uint
	Close  bool
	Code   uint
	Group  uint
	Zone   uint

	Tail string
}

func (d *DialData) Validate() error {
	if d.PP > 99 {
		return errors.New("invalid PP number")
	}
	if d.R > 9 {
		return errors.New("invalid R number")
	}

	if d.Object > 9999 {
		return errors.New("invalid object number")
	}
	if d.Code > 999 {
		return errors.New("invalid code")
	}
	if d.Group > 99 {
		return errors.New("invalid group number")
	}
	if d.Zone > 999 {
		return errors.New("invalid zone number")
	}

	return nil
}
