package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

func (this *User) Online() {
	// 加入到onlinemap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	// 广播到message
	this.server.BroadCast(this, "已上线")
}

func (this *User) Offline() {
	// 加入到onlinemap
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	// 广播到message
	this.server.BroadCast(this, "已下线")

}

func (this *User) SendMessage(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线\n"
			this.SendMessage(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 更改用户名
		newName := strings.Split(msg, "|")[1]
		if _, ok := this.server.OnlineMap[newName]; ok {
			this.SendMessage("当前用户名已存在")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMessage("您已经更新用户名: " + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式 to|张三|消息内容
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMessage("消息格式不正确")
			return
		}

		if remoteUser, ok := this.server.OnlineMap[remoteName]; !ok {
			this.SendMessage("改用户名不存在")
			return
		} else {
			content := strings.Split(msg, "|")[2]
			if content == "" {
				this.SendMessage("无消息内容，请重发\n")
				return
			}
			remoteUser.SendMessage(this.Name + "对您说: " + content)
		}

	} else {
		this.server.BroadCast(this, msg)
	}

}

// 监听 user.c 一旦有消息，就发送给对端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
