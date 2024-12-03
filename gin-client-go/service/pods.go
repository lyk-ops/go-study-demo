package service

import (
	"context"
	"errors"
	"gin-client-go/client"
	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
	"net/http"
	"sync"
)

func GetPods() ([]v1.Pod, error) {
	clientSet, err := client.GetK8sClientSet()
	if err != nil {
		klog.Fatal(err)
		return nil, err
	}
	list, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}
	return list.Items, nil

}

type WsMessage struct {
	MessageType int
	Data        []byte
}
type WsConnection struct {
	wsSocket  *websocket.Conn // websocket 连接
	inChan    chan *WsMessage // 接收消息的 chan
	outChan   chan *WsMessage // 发送消息的 chan
	mutex     sync.Mutex      // 锁, 防止并发读写
	isClosed  bool            // 标记是否关闭
	closeChan chan byte       // 用于通知连接关闭的通道
}

func (wsConn *WsConnection) wsReadLoop() {
	var (
		msgType int
		data    []byte
		msg     *WsMessage
		err     error
	)
	for {
		if msgType, data, err = wsConn.wsSocket.ReadMessage(); err != nil {
			goto ERROR
		}
		msg = &WsMessage{
			MessageType: msgType,
			Data:        data,
		}
		select {
		case wsConn.inChan <- msg:
			goto CLOSED
		}
	}
ERROR:
	wsConn.WsClose()
CLOSED:
}
func (wsConn *WsConnection) wsWriteLoop() {
	var (
		msg *WsMessage
		err error
	)
	for {
		select {
		case msg = <-wsConn.outChan:
			if err = wsConn.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				goto ERROR
			}
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.WsClose()
CLOSED:
}
func (wsConn *WsConnection) WsClose() {
	err := wsConn.wsSocket.Close()
	if err != nil {
		klog.Error("websocket close error:", err)
		return
	}
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if wsConn.isClosed {
		return
	} else {
		wsConn.isClosed = true
		close(wsConn.inChan)
	}
}

func (wsConn *WsConnection) WsWrite(messageType int, data []byte) (err error) {
	select {
	case wsConn.outChan <- &WsMessage{MessageType: messageType, Data: data}:
		return
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}
func (wsConn *WsConnection) WsRead() (msg *WsMessage, err error) {
	select {
	case msg = <-wsConn.inChan:
		return
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func InitWebSocket(resp http.ResponseWriter, req *http.Request) (wsConn *WsConnection, err error) {
	var (
		wsSocket *websocket.Conn
	)
	if wsSocket, err = wsUpgrader.Upgrade(resp, req, nil); err != nil {
		klog.Error("websocket upgrade error:", err)
		return nil, err
	}
	wsConn = &WsConnection{
		wsSocket:  wsSocket,
		inChan:    make(chan *WsMessage, 1000),
		outChan:   make(chan *WsMessage, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}
	//读取协议
	go wsConn.wsReadLoop()
	//发送协议
	go wsConn.wsWriteLoop()
	return wsConn, err
}

type streamHander struct {
	WsConn      *WsConnection
	resizeEvent chan remotecommand.TerminalSize
}

func (handler *streamHander) Write(p []byte) (size int, err error) {
	copyData := make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	err = handler.WsConn.WsWrite(websocket.TextMessage, copyData)
	return 0, err
}

type xtermMessage struct {
	MsgType string `json:"msg"`
	Input   string `json:"input"`
	Rows    uint16 `json:"rows"`
	Cols    uint16 `json:"cols"`
}

func (handler *streamHander) Read(p []byte) (size int, err error) {
	var (
		xtermMsg xtermMessage
		msg      *WsMessage
	)
	msg, err = handler.WsConn.WsRead()
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	//解析
	if err = json.Unmarshal(msg.Data, &xtermMsg); err != nil {
		klog.Error("json unmarshal error:", err)
		return 0, err
	}
	if xtermMsg.MsgType == "resize" {
		handler.resizeEvent <- remotecommand.TerminalSize{
			Width:  xtermMsg.Cols,
			Height: xtermMsg.Rows,
		}
	} else if xtermMsg.MsgType == "input" {
		size = len(p)
		copy(p, xtermMsg.Input)
	}
	return
}
func (handler *streamHander) Next() (size *remotecommand.TerminalSize) {
	ret := <-handler.resizeEvent
	size = &ret
	return
}
func WebSSH(namespace, podname, containerName, method string, resp http.ResponseWriter, req *http.Request) error {
	var (
		err      error
		executor remotecommand.Executor
		wsConn   *WsConnection
		handler  *streamHander
	)
	config, err := client.GetRestConfig()
	if err != nil {
		klog.Fatal(err)
		return err
	}
	clientSet, err := client.GetK8sClientSet()
	if err != nil {
		klog.Fatal(err)
		return err
	}
	reqSSH := clientSet.CoreV1().RESTClient().Post().Resource("pods").Namespace(namespace).Name(podname).SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: containerName,
			Command:   []string{"/bin/sh", "-c", method},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	if executor, err = remotecommand.NewSPDYExecutor(config, "POST", reqSSH.URL()); err != nil {
		klog.Fatal(err)
		return err
	}
	if wsConn, err = InitWebSocket(resp, req); err != nil {
		return err
	}
	handler = &streamHander{WsConn: wsConn, resizeEvent: make(chan remotecommand.TerminalSize)}
	if err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		Tty:               false,
		TerminalSizeQueue: handler,
	}); err != nil {
		goto END
	}
	return err
END:
	klog.Errorln(err)
	wsConn.WsClose()
	return err
}
