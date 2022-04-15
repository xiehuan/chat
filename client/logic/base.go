package logic

type BaseMsg struct {
	UserName string
	Content  string
}

type PushMsg struct {
	PushConnMsg []*BaseMsg
}
