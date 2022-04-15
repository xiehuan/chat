package logic

import "log"

type Service struct {
	connManage *ConnManage
	roomManage *RoomManage
	userManage *UserManage
	msgManage  *MsgManage
}

func (s *Service) Start() {
	// 初始化聊天室
	roomManage := &RoomManage{}
	roomManage.Start(s)
	s.roomManage = roomManage
	log.Printf("roomManage begin")

	// 初始化用户信息
	userManage := &UserManage{}
	userManage.Start(s)
	s.userManage = userManage
	log.Printf("userManage begin")

	// 初始化msg
	msgManage := &MsgManage{}
	msgManage.Start(s)
	s.msgManage = msgManage
	log.Printf("msgManage begin")

	// 初始化连接
	connManage := &ConnManage{}
	connManage.Start(s)
	s.connManage = connManage
	log.Printf("connManage begin")
}

func (s *Service) Stop() {
	s.connManage.Stop()
	s.msgManage.Stop()
	s.userManage.Stop()
	s.roomManage.Stop()
}
