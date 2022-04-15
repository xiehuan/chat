package logic

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn

	writeChan chan []byte

	wg        sync.WaitGroup
	closeChan chan bool
}

func NewClient() *Client {
	return &Client{
		closeChan: make(chan bool, 1),
		writeChan: make(chan []byte, 1024),
	}
}

func (c *Client) CreateConn() {
	conn, err := net.Dial("tcp", "127.0.0.1:5678")
	if err != nil {
		log.Fatalf("dial err %s", err.Error())
		return
	}
	c.conn = conn

	// 读协程
	go c.connRead()

	// 写协程
	go c.connWrite()
}

func (c *Client) connRead() {
	buffer := make([]byte, 10240)
	for {
		n, err := c.conn.Read(buffer)
		if err != nil {
			log.Printf("conn read buffer err %s", err.Error())
			continue
		}
		bytesMsg := buffer[:n]

		pushMsg := &PushMsg{}
		err = json.Unmarshal(bytesMsg, pushMsg)
		if err != nil {
			log.Printf("unmarshal msg err %s", err.Error())
			continue
		}

		for _, baseMsg := range pushMsg.PushConnMsg {
			content := ""
			if baseMsg.UserName != "" {
				content += baseMsg.UserName + ":"
			}
			content += baseMsg.Content

			log.Println(content)
		}
	}
}

func (c *Client) connWrite() {
	for {
		select {
		case msg := <-c.writeChan:
			_, err := c.conn.Write(msg)
			if err != nil {
				log.Printf("write msg %s err %s", msg, err.Error())
				continue
			}
		case <-c.closeChan:
			return
		}
	}
}

func (c *Client) ReadStdin() {
	// 用户教程
	log.Printf("use \"/name xxx\" to login")
	log.Printf("use \"/room num\" to choose room, room num 0 - 9")
	log.Printf("use \"/stats name\" to show user info")
	log.Printf("use \"/popular roomNum\" to get most popular word in 10 min, room num 0 - 9")

	reader := bufio.NewReader(os.Stdin)
	for {
		s, _ := reader.ReadString('\n')
		sysType := runtime.GOOS
		if sysType == "windows" {
			s = strings.TrimRight(s, "\r\n")
		} else if sysType == "linux" {
			s = strings.TrimRight(s, "\n")
		}
		// 发给conn，写给服务器
		c.writeChan <- []byte(s)
	}
}
