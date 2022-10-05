package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       9999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return client
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>请输入合法范围的数字<<<")
		return false
	}
}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>请输入聊天内容, exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>请输入聊天内容, exit退出.")
		fmt.Scanln(&chatMsg)
	}

}

func (client *Client) SelectUsers() {
	snedMsg := "who\n"
	_, err := client.conn.Write([]byte(snedMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println(">>>>请输入聊天对象[用户名]， exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>请输入聊天内容")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>请输入聊天内容")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		remoteName = ""
		fmt.Println(">>>>请输入聊天对象[用户名]， exit退出")
		fmt.Scanln(&remoteName)

	}

}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) DealResponse() {
	// for {
	// 	buf := make([]byte, 4096)
	// 	client.conn.Read(buf)
	// 	fmt.Println(buf)
	// }

	// 以上步骤的简写形式, 永久阻塞监听
	io.Copy(os.Stdout, client.conn)

}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		switch client.flag {
		case 1:
			// 公聊
			fmt.Println("公聊模式选择...")
			client.PublicChat()
			break
		case 2:
			// 私聊
			fmt.Println("私聊模式选择...")
			break
		case 3:
			// 更新用户名
			fmt.Println("更新用户名选择...")
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 9999, "设置服务器端口(默认是9999)")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> 链接服务器失败...")
		return
	}

	go client.DealResponse()

	fmt.Println(">>>>> 链接服务器成功...")

	// 启东客户端的业务
	client.Run()

	select {}

}
