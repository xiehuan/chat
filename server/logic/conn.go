package logic

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
)

type ConnManage struct {
	s *Service

	UserConn map[int]*UserConn

	listener net.Listener
	connNum  int
}

type UserConn struct {
	ConnID   int
	UserName string

	conn        net.Conn
	receiveChan chan *ConnMsg
	sendChan    chan *PushConnMsg
	closeChan   chan bool
}

func (cm *ConnManage) init(s *Service) {
	cm.s = s
	cm.connNum = 0
	cm.UserConn = make(map[int]*UserConn)
}

func (cm *ConnManage) Start(s *Service) {
	cm.init(s)

	// 初始化监听
	listener, err := net.Listen("tcp", "127.0.0.1:5678")
	if err != nil {
		log.Printf("listen port err %s", err.Error())
		return
	}
	cm.listener = listener
	// 监听
	go cm.listen()
}

func (cm *ConnManage) Stop() {
	// 关闭监听
	cm.listener.Close()

	// 关闭用户连接
	for _, user := range cm.UserConn {
		user.Stop()
	}
}

func (cm *ConnManage) listen() {
	for {
		conn, err := cm.listener.Accept()
		if err != nil {
			log.Printf("listener accept err %s", err.Error())
			continue
		}
		cm.connNum++
		cm.newUserConn(cm.connNum, conn)
	}
}

func (cm *ConnManage) newUserConn(id int, conn net.Conn) {
	// 初始化一个连接
	userConn := &UserConn{
		ConnID:      id,
		conn:        conn,
		receiveChan: cm.s.msgManage.receiveMsgChan,
		sendChan:    make(chan *PushConnMsg, 64),
		closeChan:   make(chan bool, 1),
	}
	cm.UserConn[id] = userConn

	// 注册用户的接收，发送的chan
	cm.s.msgManage.connMsgDealChan <- &ConnChanMsg{
		ConnID:   id,
		SendChan: userConn.sendChan,
	}

	log.Printf("new conn %d build", id)

	// 启动读写协程
	userConn.Start()
}

func (uc *UserConn) Start() {
	// 读消息
	go uc.connRead()

	// 写消息
	go uc.connWrite()
}

func (uc *UserConn) Stop() {
	uc.conn.Close()
	close(uc.closeChan)
}

func (uc *UserConn) connRead() {
	buffer := make([]byte, 1024)
	for {
		n, err := uc.conn.Read(buffer)
		// 如果报错，做回收处理
		if err != nil && err != io.EOF {
			log.Printf("conn %d read buffer err %s", uc.ConnID, err.Error())

			// 通知下线
			msg := &ConnMsg{
				ConnID:  uc.ConnID,
				Content: fmt.Sprintf("%s %d", Logout, uc.ConnID),
			}
			uc.receiveChan <- msg
			return
		}

		buffMsg := string(buffer[:n])
		msg := &ConnMsg{
			ConnID:  uc.ConnID,
			Content: buffMsg,
		}
		uc.receiveChan <- msg

		log.Printf("receive msg %s", buffMsg)
	}
}

func (uc *UserConn) connWrite() {
	for {
		select {
		case msg := <-uc.sendChan:
			jsonBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("msg %v marshal err %s", msg, err.Error())
				continue
			}
			log.Printf("server write msg %s", jsonBytes)
			_, err = uc.conn.Write(jsonBytes)
			if err != nil {
				log.Printf("conn %d write msg %v err %s", uc.ConnID, msg, err.Error())
				continue
			}
		case <-uc.closeChan:
			return
		}
	}
}
