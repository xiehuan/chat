package logic

type ConnMsg struct {
	ConnID  int
	Content string
}

type BaseMsg struct {
	UserName string
	Content  string
}

type PushMsg struct {
	ConnID  []int
	PushMsg []*BaseMsg
}

type PushConnMsg struct {
	PushConnMsg []*BaseMsg
}

type ConnChanMsg struct {
	ConnID   int
	SendChan chan *PushConnMsg
}

type UserStatsMsg struct {
	ConnID int
	Name   string
}

type UserLogoutMsg struct {
	ConnID int
}

type UserNameMsg struct {
	ConnID int
	Name   string
}

type UserRoomMsg struct {
	RoomID int
	ConnID int
}

type UserSendMsg struct {
	ConnID  int
	Content string
}

type RoomChangeMsg struct {
	OldRoomID int
	NewRoomID int
	ConnID    int
}

type RoomReceiveMsg struct {
	ConnID   int
	UserName string
	RoomID   int
	Content  string
}

type RoomLogoutMsg struct {
	ConnID int
	RoomID int
}

type RoomPopularMsg struct {
	ConnID int
	RoomID int
}
