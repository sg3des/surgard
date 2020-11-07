package surgard

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

const cmdLayout = "5%02d%01d 18%04d%s%03d%02d%03d"
const eoc = "\x14"

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

	log.Printf("surgard: %s", cmd)

	if _, err := conn.Write([]byte(cmd + eoc)); err != nil {
		return fmt.Errorf("failed to send re-command, %v", err)
	}

	// read response, should be 1 byte, 0x06 - command recieve, 0x15 - need to repeat
	b := make([]byte, 1)
	if _, err := conn.Read(b); err != nil {
		return fmt.Errorf("failed to read response, %v", err)
	}

	log.Printf("%02x", b)

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

	cmd := fmt.Sprintf(cmdLayout, d.PP, d.R, d.Object, q, d.Code, d.Group, d.Zone)
	return c.Send(cmd)
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
