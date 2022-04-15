# chat

## 简介

IDE：windows + goland2021.3.4 + go1.16

chat下分为客户端（client）目录和服务器（server）目录

server实现：

conn.go 监听端口，和客户端建立连接

room.go 房间相关逻辑

user.go 用户相关逻辑

msg.go 消息中转

client实现：

client.go 和服务器建立连接，收发消息并展示


# 编译运行

## linux

1.编译服务器

cd ./server

go build -o server main.go

./server

注意保持脏词库list.txt和可执行文件在同一目录下

2.编译客户端

cd ./server

go build -o client main.go

./client

## windows

1.编译服务器

cd ./server

go build -o server.exe main.go

./server.exe

注意保持脏词库list.txt和可执行文件在同一目录下

2.编译客户端

cd ./server

go build -o client.exe main.go

./client.exe

# 使用

命令：

/name xxx 建号或者登录

/room num 选择聊天房间0-9

/stats xxx 某用户的状态

/popular num 某个房间（0-9）十分钟内出现频率最大的词

流程：

1.启动server

2.启动client，必须先执行/name进行登录，/room选择聊天房间方可进行聊天，否则聊天无效

