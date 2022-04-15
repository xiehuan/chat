package logic

import (
	"log"
	"strings"
	"sync"
	"time"
)

type RoomManage struct {
	s *Service

	Rooms map[int]*Room

	roomChangeChan     chan *RoomChangeMsg  // 切换房间
	roomReceiveMsgChan chan *RoomReceiveMsg // 房间聊天
	roomPopularChan    chan *RoomPopularMsg // 房间十分钟内最高频率单词
	roomLogoutMsg      chan *RoomLogoutMsg  // 登出用户

	wg        sync.WaitGroup
	closeChan chan bool
}

type Room struct {
	RoomID  int         // 房间唯一ID
	ChatMsg []*ChatMsg  // 房间内消息
	Users   map[int]int // 房间内玩家
}

type ChatMsg struct {
	UserName   string
	MsgContent string
	MsgTime    int64
}

func (rm *RoomManage) init(s *Service) {
	rm.s = s
	rm.Rooms = make(map[int]*Room)
	rm.roomChangeChan = make(chan *RoomChangeMsg, 128)
	rm.roomReceiveMsgChan = make(chan *RoomReceiveMsg, 1024)
	rm.roomPopularChan = make(chan *RoomPopularMsg, 1024)
	rm.roomLogoutMsg = make(chan *RoomLogoutMsg, 64)
	rm.closeChan = make(chan bool, 1)
}

func (rm *RoomManage) Start(s *Service) {
	rm.init(s)

	rm.initRoom()

	rm.wg.Add(1)
	go rm.roomLogic()
}

func (rm *RoomManage) Stop() {
	close(rm.closeChan)
	rm.wg.Wait()
}

func (rm *RoomManage) initRoom() {
	for i := 0; i < RoomNum; i++ {
		room := &Room{
			RoomID:  i,
			ChatMsg: make([]*ChatMsg, 0),
			Users:   make(map[int]int),
		}
		rm.Rooms[i] = room
		log.Printf("init chat room %d", i)
	}
}

func (rm *RoomManage) roomLogic() {
	defer rm.wg.Done()
	for {
		select {
		case msg := <-rm.roomReceiveMsgChan:
			rm.roomMsgLogic(msg)
		case roomChangeMsg := <-rm.roomChangeChan:
			rm.roomChangeLogic(roomChangeMsg)
		case roomPopularMsg := <-rm.roomPopularChan:
			rm.roomPopularLogic(roomPopularMsg)
		case roomLogoutMsg := <-rm.roomLogoutMsg:
			rm.roomLogoutLogic(roomLogoutMsg)
		case <-rm.closeChan:
			return
		}
	}
}

func (rm *RoomManage) roomMsgLogic(msg *RoomReceiveMsg) {
	room := rm.Rooms[msg.RoomID]
	if room == nil {
		return
	}

	// 转发给房间内所有人
	connIDs := make([]int, 0)
	for connID := range room.Users {
		connIDs = append(connIDs, connID)
	}
	pushToOtherMsg := make([]*BaseMsg, 0)
	pushToOtherMsg = append(pushToOtherMsg, &BaseMsg{
		UserName: msg.UserName,
		Content:  msg.Content,
	})
	pushMsg := &PushMsg{
		ConnID:  connIDs,
		PushMsg: pushToOtherMsg,
	}
	rm.s.msgManage.pushMsgChan <- pushMsg

	// 记录此条消息
	now := time.Now().Unix()
	chatMsg := &ChatMsg{
		UserName:   msg.UserName,
		MsgContent: msg.Content,
		MsgTime:    now,
	}
	room.ChatMsg = append(room.ChatMsg, chatMsg)

	// TODO 消息清理 十分钟之前并且消息不处于最近50条
}

func (rm *RoomManage) roomChangeLogic(msg *RoomChangeMsg) {
	// 找到旧的房间
	oldRoom := rm.Rooms[msg.OldRoomID]
	if oldRoom != nil {
		oldRoom.delUser(msg.ConnID)
	}

	// 加入新的房间
	newRoom := rm.Rooms[msg.NewRoomID]
	if newRoom != nil {
		newRoom.addUser(msg.ConnID)
	}

	if newRoom == nil {
		return
	}

	// 找出房间最近50条
	roomMsg := make([]*ChatMsg, 0)
	roomMsgLen := len(newRoom.ChatMsg)
	if roomMsgLen > JoinRoomChatMsg {
		roomMsg = newRoom.ChatMsg[roomMsgLen-JoinRoomChatMsg:]
	} else {
		roomMsg = newRoom.ChatMsg
	}

	// 房间内没消息不推送
	if len(roomMsg) == 0 {
		return
	}

	// 推送消息
	pushToOtherMsg := make([]*BaseMsg, 0)
	for _, cMsg := range roomMsg {
		pushToOtherMsg = append(pushToOtherMsg, &BaseMsg{
			UserName: cMsg.UserName,
			Content:  cMsg.MsgContent,
		})
	}
	connIDs := make([]int, 0)
	connIDs = append(connIDs, msg.ConnID)
	pushMsg := &PushMsg{
		ConnID:  connIDs,
		PushMsg: pushToOtherMsg,
	}
	rm.s.msgManage.pushMsgChan <- pushMsg

}

func (r *Room) delUser(connID int) {
	delete(r.Users, connID)
}

func (r *Room) addUser(connID int) {
	r.Users[connID] = connID
}

func (rm *RoomManage) roomPopularLogic(msg *RoomPopularMsg) {
	room := rm.Rooms[msg.RoomID]
	if room == nil {
		return
	}

	// 获取最多频率单词
	maxPopularWord := getMaxPopularWord(room.ChatMsg)

	// 发送给用户
	connIDs := make([]int, 0)
	connIDs = append(connIDs, msg.ConnID)
	pushToOtherMsg := make([]*BaseMsg, 0)
	pushToOtherMsg = append(pushToOtherMsg, &BaseMsg{
		Content: maxPopularWord,
	})
	pushMsg := &PushMsg{
		ConnID:  connIDs,
		PushMsg: pushToOtherMsg,
	}
	rm.s.msgManage.pushMsgChan <- pushMsg
}

func (rm *RoomManage) roomLogoutLogic(msg *RoomLogoutMsg) {
	room := rm.Rooms[msg.RoomID]
	if room == nil {
		return
	}
	room.delUser(msg.ConnID)
}

func getMaxPopularWord(chatMsg []*ChatMsg) string {
	// 找到十分钟节点
	now := time.Now().Unix()
	lastTenMinTime := now - PopularBeforeSecond
	wordCount := make(map[string]int)
	for _, cMsg := range chatMsg {
		if cMsg.MsgTime < lastTenMinTime {
			continue
		}

		// 对消息处理
		msgArr := strings.Split(cMsg.MsgContent, " ")
		for _, m := range msgArr {
			wordCount[m]++
		}
	}

	// 找到出现次数最多的词
	maxCount := 0
	maxWord := ""
	for word, count := range wordCount {
		if count > maxCount {
			maxWord = word
			maxCount = count
		}
	}

	return maxWord
}
