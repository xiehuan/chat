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
