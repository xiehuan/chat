package logic

import (
	"testing"
	"time"
)

func Test_getMaxPopularWord(t *testing.T) {
	type args struct {
		chatMsg []*ChatMsg
	}
	now := time.Now().Unix()
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test_freq",
			args{
				chatMsg: []*ChatMsg{
					{
						"",
						"aa bb cc dd ee",
						now,
					},
					{
						"",
						"aa aa cc dd ee",
						now,
					},
				},
			},
			"aa",
		},
		{
			"test_time",
			args{
				chatMsg: []*ChatMsg{
					{
						"",
						"aa bb cc dd",
						now,
					},
					{
						"",
						"aa aa aa aa",
						now - 601,
					},
					{
						"",
						"bb bb cc dd",
						now,
					},
				},
			},
			"bb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMaxPopularWord(tt.args.chatMsg); got != tt.want {
				t.Errorf("getMaxPopularWord() = %v, want %v", got, tt.want)
			}
		})
	}
}
