package logic

import (
	"log"
	"strconv"
	"strings"
	"sync"
)

type MsgManage struct {
	s *Service

	receiveMsgChan chan *ConnMsg // 接收来自所有玩家的消息
	pushMsgChan    chan *PushMsg // 推送给玩家的消息

	connMsgDealChan chan *ConnChanMsg         // 注册conn对应的channel
	sendUserMsgChan map[int]chan *PushConnMsg // 发送给conn的channel

	wg        sync.WaitGroup
	closeChan chan bool
}

func (mm *MsgManage) init(s *Service) {
	mm.s = s
	mm.receiveMsgChan = make(chan *ConnMsg, 1024)
	mm.pushMsgChan = make(chan *PushMsg, 1024)
	mm.sendUserMsgChan = make(map[int]chan *PushConnMsg)
	mm.connMsgDealChan = make(chan *ConnChanMsg, 1024)
	mm.closeChan = make(chan bool, 1)
}
func (mm *MsgManage) Start(s *Service) {
	mm.init(s)

	mm.wg.Add(1)
	go mm.msgDeal()
}

func (mm *MsgManage) Stop() {
	close(mm.closeChan)
	mm.wg.Wait()
}

func (mm *MsgManage) msgDeal() {
	defer mm.wg.Done()
	for {
		select {
		case connChanMsg := <-mm.connMsgDealChan:
			// 注册新连接的消息通道
			mm.sendUserMsgChan[connChanMsg.ConnID] = connChanMsg.SendChan
		case receiveMsg := <-mm.receiveMsgChan:
			// 根据收到的消息做不同处理
			mm.msgLogic(receiveMsg)
		case pushMsg := <-mm.pushMsgChan:
			// 推送给客户端的消息
			mm.pushMsgToConn(pushMsg)
		case <-mm.closeChan:
			return
		}
	}
}

func (mm *MsgManage) pushMsgToConn(msg *PushMsg) {
	connMsg := &PushConnMsg{
		PushConnMsg: msg.PushMsg,
	}
	log.Printf("send to user %v msg %v", msg.ConnID, msg.PushMsg)
	// 发送给对应的玩家
	for _, connID := range msg.ConnID {
		if sendChan, ok := mm.sendUserMsgChan[connID]; ok {
			sendChan <- connMsg
		}
	}
}

func (mm *MsgManage) msgLogic(msg *ConnMsg) {
	// 检查是否是特殊处理消息
	isGM := mm.msgCommand(msg)
	if isGM {
		return
	}

	// 直接发给user处理
	userSendMsg := &UserSendMsg{
		ConnID:  msg.ConnID,
		Content: msg.Content,
	}
	mm.s.userManage.userSendMsgChan <- userSendMsg
}

func (mm *MsgManage) msgCommand(msg *ConnMsg) bool {
	msgArr := strings.Split(msg.Content, " ")
	if len(msgArr) < 2 {
		return false
	}
	switch msgArr[0] {
	// 获取用户信息
	case Stats:
		userName := msgArr[1]
		if userName == "" {
			return true
		}
		userStatMsg := &UserStatsMsg{
			Name:   userName,
			ConnID: msg.ConnID,
		}
		mm.s.userManage.userStatMsgChan <- userStatMsg
		return true
	// 获取房间内最高频率单词
	case Popular:
		roomID, err := strconv.Atoi(msgArr[1])
		if err != nil {
			log.Printf("popular room id %s atoi err %s", msgArr[1], err.Error())
			return true
		}
		roomPopularMsg := &RoomPopularMsg{
			RoomID: roomID,
			ConnID: msg.ConnID,
		}
		mm.s.roomManage.roomPopularChan <- roomPopularMsg
		return true
	// 取名
	case Name:
		userName := msgArr[1]
		if userName == "" {
			return true
		}

		userNameMsg := &UserNameMsg{
			Name:   userName,
			ConnID: msg.ConnID,
		}
		mm.s.userManage.userNameMsgChan <- userNameMsg
		return true
	// 切换房间
	case ChangeRoom:
		roomID, err := strconv.Atoi(msgArr[1])
		if err != nil {
			log.Printf("choose room id %s atoi err %s", msgArr[1], err.Error())
			mm.sendToUserMsg(msg.ConnID, RoomIDErr)
			return true
		}

		// roomID越界
		if roomID < 0 || roomID > RoomNum-1 {
			mm.sendToUserMsg(msg.ConnID, RoomIDErr)
			return true
		}

		userRoomMsg := &UserRoomMsg{
			RoomID: roomID,
			ConnID: msg.ConnID,
		}
		mm.s.userManage.userRoomMsgChan <- userRoomMsg
		return true
	// 登出
	case Logout:
		userMsg := &UserLogoutMsg{
			ConnID: msg.ConnID,
		}
		mm.s.userManage.userLogoutMsgChan <- userMsg
	}

	return false
}

func (mm *MsgManage) sendToUserMsg(connID int, content string) {
	connIDs := make([]int, 0)
	connIDs = append(connIDs, connID)
	pushMsg := make([]*BaseMsg, 0)
	pushMsg = append(pushMsg, &BaseMsg{
		Content: content,
	})
	mm.pushMsgChan <- &PushMsg{
		ConnID:  connIDs,
		PushMsg: pushMsg,
	}
}
