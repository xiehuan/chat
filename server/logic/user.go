package logic

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type UserManage struct {
	s *Service

	badWords []string

	users            map[string]*User
	userConnIDToName map[int]string

	userNameMsgChan   chan *UserNameMsg   // 取名
	userRoomMsgChan   chan *UserRoomMsg   // 选择房间
	userSendMsgChan   chan *UserSendMsg   // 聊天消息
	userStatMsgChan   chan *UserStatsMsg  // 用户状态
	userLogoutMsgChan chan *UserLogoutMsg // 用户登出

	wg        sync.WaitGroup
	closeChan chan bool
}

type User struct {
	Name       string
	LoginTime  int64
	LogoutTime int64
	OnlineTime int64
	RoomID     int
	ConnID     int
	Status     int
}

func (um *UserManage) init(s *Service) {
	um.s = s
	um.users = make(map[string]*User)
	um.badWords = make([]string, 0)
	um.userConnIDToName = make(map[int]string)
	um.userNameMsgChan = make(chan *UserNameMsg, 64)
	um.userRoomMsgChan = make(chan *UserRoomMsg, 64)
	um.userSendMsgChan = make(chan *UserSendMsg, 1024)
	um.userStatMsgChan = make(chan *UserStatsMsg, 64)
	um.userLogoutMsgChan = make(chan *UserLogoutMsg, 64)
	um.closeChan = make(chan bool, 1)
}

func (um *UserManage) Start(s *Service) {
	um.init(s)

	// 读脏词库
	um.loadBadWords()

	// 启动
	um.wg.Add(1)
	go um.userLogic()
}

func (um *UserManage) Stop() {
	close(um.closeChan)
	um.wg.Wait()
}

func (um *UserManage) loadBadWords() {
	listFile, err := os.Open("list.txt")
	if err != nil {
		log.Printf("load bad words err %s", err.Error())
		return
	}
	defer listFile.Close()

	br := bufio.NewReader(listFile)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}

		um.badWords = append(um.badWords, string(line))
	}
}

func (um *UserManage) userLogic() {
	defer um.wg.Done()
	for {
		select {
		case nameMsg := <-um.userNameMsgChan:
			um.nameLogic(nameMsg)
		case roomMsg := <-um.userRoomMsgChan:
			um.chooseRoomLogic(roomMsg)
		case sendMsg := <-um.userSendMsgChan:
			um.sendMsgLogic(sendMsg)
		case statMsg := <-um.userStatMsgChan:
			um.statLogic(statMsg)
		case logoutMsg := <-um.userLogoutMsgChan:
			um.logoutLogic(logoutMsg)
		case <-um.closeChan:
			return
		}
	}

}

func (um *UserManage) nameLogic(msg *UserNameMsg) {
	userName := um.userConnIDToName[msg.ConnID]
	if userName != "" {
		um.sendSingleMsg(msg.ConnID, AlreadyLogin)
		return
	}

	user := um.users[msg.Name]
	if user != nil && user.Status == StatusOnline {
		um.sendSingleMsg(msg.ConnID, NameRepeat)
		return
	}

	if user == nil {
		user = &User{
			Name: msg.Name,
		}
		um.users[msg.Name] = user
	}
	user.ConnID = msg.ConnID
	user.LoginTime = time.Now().Unix()
	user.Status = StatusOnline
	um.userConnIDToName[msg.ConnID] = msg.Name

	// 发消息，登录成功
	um.sendSingleMsg(msg.ConnID, LoginSuccess)
}

func (um *UserManage) chooseRoomLogic(msg *UserRoomMsg) {
	userName := um.userConnIDToName[msg.ConnID]
	if userName == "" {
		return
	}
	user := um.users[userName]
	if user == nil {
		return
	}

	// 如果切换的房间相同
	if user.RoomID == msg.RoomID {
		return
	}

	lastRoomID := user.RoomID
	user.RoomID = msg.RoomID

	// 发消息，选择房间成功
	connIDs := make([]int, 0)
	connIDs = append(connIDs, msg.ConnID)
	pushToOtherMsg := make([]*BaseMsg, 0)
	pushToOtherMsg = append(pushToOtherMsg, &BaseMsg{
		UserName: "",
		Content:  JoinRoomSuccess,
	})
	pushMsg := &PushMsg{
		ConnID:  connIDs,
		PushMsg: pushToOtherMsg,
	}
	um.s.msgManage.pushMsgChan <- pushMsg

	// 通知房间管理
	roomChangeMsg := &RoomChangeMsg{
		OldRoomID: lastRoomID,
		NewRoomID: msg.RoomID,
		ConnID:    user.ConnID,
	}
	um.s.roomManage.roomChangeChan <- roomChangeMsg
}

func (um *UserManage) sendMsgLogic(msg *UserSendMsg) {
	userName := um.userConnIDToName[msg.ConnID]
	if userName == "" {
		return
	}
	user := um.users[userName]
	if user == nil {
		return
	}
	if user.RoomID == 0 {
		return
	}

	// 单词过滤
	msgContent := msg.Content
	for _, word := range um.badWords {
		msgContent = strings.ReplaceAll(msgContent, word, "*")
	}

	// 发消息给房间
	roomMsg := &RoomReceiveMsg{
		ConnID:   user.ConnID,
		UserName: userName,
		RoomID:   user.RoomID,
		Content:  msgContent,
	}
	um.s.roomManage.roomReceiveMsgChan <- roomMsg
}

func (um *UserManage) statLogic(msg *UserStatsMsg) {
	user := um.users[msg.Name]
	if user == nil {
		return
	}

	// 发消息
	onlineTime := user.OnlineTime
	if user.LogoutTime < user.LoginTime {
		onlineTime += time.Now().Unix() - user.LoginTime
	}
	content := fmt.Sprintf("loginTime:%d onlineTime:%d roomID:%d", user.LoginTime, onlineTime, user.RoomID)
	um.sendSingleMsg(msg.ConnID, content)
}

func (um *UserManage) logoutLogic(msg *UserLogoutMsg) {
	userName := um.userConnIDToName[msg.ConnID]
	if userName == "" {
		return
	}
	user := um.users[userName]
	if user == nil {
		return
	}

	now := time.Now().Unix()
	user.OnlineTime += now - user.LoginTime
	user.LogoutTime = now
	user.Status = StatusLogout
	user.RoomID = 0

	// 通知房间
	roomMsg := &RoomLogoutMsg{
		ConnID: user.ConnID,
		RoomID: user.RoomID,
	}
	um.s.roomManage.roomLogoutMsg <- roomMsg
}

func (um *UserManage) sendSingleMsg(connID int, content string) {
	connIDs := make([]int, 0)
	connIDs = append(connIDs, connID)
	pushToOtherMsg := make([]*BaseMsg, 0)
	pushToOtherMsg = append(pushToOtherMsg, &BaseMsg{
		UserName: "",
		Content:  content,
	})
	pushMsg := &PushMsg{
		ConnID:  connIDs,
		PushMsg: pushToOtherMsg,
	}
	um.s.msgManage.pushMsgChan <- pushMsg
}
