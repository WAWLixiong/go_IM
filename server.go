package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return &server
}

// 监听message广播消息，并广播给user

func (this *Server) ListenMessaged() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, user := range this.OnlineMap {
			user.C <- msg
		}
		this.mapLock.Unlock()
	}

}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	// ... 当前链接的业务
	fmt.Println("conn success")

	user := NewUser(conn, this)
	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn read err:", err)
				return
			}
			// 提取用户消息，去掉\n
			msg := string(buf[:n-1])
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户活跃
			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
			// 什么也不做, 退出一次for循环, 用户的定时器就会刷新
		case <-time.After(time.Minute * 10):
			// 可读，则说明已经超时
			user.SendMessage("你被踢了")
			close(user.C)
			conn.Close()
			return // runtime.Goexit()
		}
	}

	// 当前的handler阻塞
	// select {}

}

func (this *Server) Start() {

	// listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// message广播到onlinemap
	go this.ListenMessaged()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listener accpet err:", err)
			continue
		}
		// do handler
		go this.Handler(conn)
	}

}
