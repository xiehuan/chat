package logic

const JoinRoomChatMsg = 50

const RoomNum = 10

const (
	JoinRoomSuccess = "Notify:JoinRoomSuccess"
	LoginSuccess    = "Notify:LoginSuccess"
	NameRepeat      = "Notify:NameRepeat"
	AlreadyLogin    = "Notify:AlreadyLogin"
	RoomIDErr       = "Notify:RoomIDErr"
)

const (
	Stats      = "/stats"
	Popular    = "/popular"
	Name       = "/name"
	ChangeRoom = "/room"
	Logout     = "/logout"
)

const (
	PopularBeforeSecond = 600
)

const (
	_ = iota
	StatusOnline
	StatusLogout
)
