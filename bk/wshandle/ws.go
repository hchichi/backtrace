package wshandle

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WsConnection struct {
	Conn         *websocket.Conn
	MsgSendCh    chan string
	MsgReceiveCh chan string
	ConnMux      sync.Mutex
	Connected    bool
}

var (
	wsConn     *WsConnection
	wsConnOnce sync.Once
	wsConnErr  error
)

func initWsConn() {
	wsConn = &WsConnection{
		MsgSendCh:    make(chan string),
		MsgReceiveCh: make(chan string),
		Connected:    false,
	}

	// 初始化连接
	if err := connect(); err != nil {
		wsConnErr = err
		return
	}

	// 启动发送协程
	go func() {
		for {
			msg := <-wsConn.MsgSendCh
			if !wsConn.Connected {
				// 尝试重连
				if err := connect(); err != nil {
					continue
				}
			}
			err := wsConn.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				wsConn.Connected = false
				// 尝试重连
				_ = connect()
			}
		}
	}()

	// 启动接收协程
	go func() {
		for {
			if !wsConn.Connected {
				// 尝试重连
				if err := connect(); err != nil {
					time.Sleep(time.Second)
					continue
				}
			}
			_, message, err := wsConn.Conn.ReadMessage()
			if err != nil {
				wsConn.Connected = false
				continue
			}
			wsConn.MsgReceiveCh <- string(message)
		}
	}()
}

func connect() error {
	if wsConn.Conn != nil {
		wsConn.Conn.Close()
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	c, _, err := dialer.Dial("wss://api.leo.moe/trace", nil)
	if err != nil {
		return err
	}

	wsConn.ConnMux.Lock()
	wsConn.Conn = c
	wsConn.Connected = true
	wsConn.ConnMux.Unlock()

	return nil
}

func GetWsConn() (*WsConnection, error) {
	wsConnOnce.Do(initWsConn)
	if wsConnErr != nil {
		return nil, wsConnErr
	}
	if !wsConn.Connected {
		return nil, errors.New("WebSocket not connected")
	}
	return wsConn, nil
}
