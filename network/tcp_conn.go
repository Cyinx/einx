package network

import (
	"errors"
	"github.com/Cyinx/einx/agent"
	"github.com/Cyinx/einx/queue"
	"github.com/Cyinx/einx/slog"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type TcpConn struct {
	agentID      AgentID
	conn         net.Conn
	closeFlag    uint32
	writeQueue   *queue.CondQueue
	serveHandler SessionHandler
	lastPingTick int64
	remoteAddr   string
	connType     int16
	pingClose    int32
	userType     interface{}

	recvBuf       *BytesBuffer
	writeBuf      *BytesBuffer
	msgPacket     transPacket
	recvCheckTime int64
	msgRecvCount  int64
	option        *TransportOption
}

func newTcpConn(raw_conn net.Conn, h SessionHandler, conn_type int16, opt *TransportOption) *TcpConn {
	nowTime := UnixTS()
	tcpAgent := &TcpConn{
		conn:         raw_conn,
		closeFlag:    0,
		writeQueue:   queue.NewCondQueue(),
		agentID:      agent.GenAgentID(),
		serveHandler: h,
		remoteAddr:   raw_conn.RemoteAddr().(*net.TCPAddr).String(),
		connType:     conn_type,
		userType:     0,

		recvBuf:       bufferPool.Get().(*BytesBuffer),
		writeBuf:      bufferPool.Get().(*BytesBuffer),
		option:        opt,
		lastPingTick:  nowTime,
		recvCheckTime: nowTime,
	}
	return tcpAgent
}

func (n *TcpConn) GetID() AgentID {
	return n.agentID
}

func (n *TcpConn) GetType() int16 {
	return n.connType
}

func (n *TcpConn) GetUserType() interface{} {
	return n.userType
}

func (n *TcpConn) SetUserType(t interface{}) {
	n.userType = t
}

func (n *TcpConn) ReadMsg() ([]byte, error) {
	return nil, nil
}

func (n *TcpConn) IsClosed() bool {
	close_flag := atomic.LoadUint32(&n.closeFlag)
	return close_flag == 1
}

func (n *TcpConn) doPushWrite(wrapper ITransportMsg) bool {
	n.writeQueue.Push(wrapper)
	return true
}

func (n *TcpConn) MultipleMsg() ITranMsgMultiple {
	x := &TransportMultiple{}
	x.trans = n
	return x
}

var writePool *sync.Pool = &sync.Pool{New: func() interface{} { return new(TransportMsgPack) }}

func (n *TcpConn) WriteMsg(msgID ProtoTypeID, b []byte) bool {
	if n.IsClosed() == true {
		return false
	}

	w := writePool.Get().(*TransportMsgPack)
	w.msgType = 'P'
	w.msgID = msgID
	w.Buf = b

	return n.doPushWrite(w)
}

func (n *TcpConn) RpcCall(msgID ProtoTypeID, b []byte) bool {
	if n.IsClosed() == true {
		return false
	}

	w := writePool.Get().(*TransportMsgPack)
	w.msgType = 'R'
	w.msgID = msgID
	w.Buf = b

	return n.doPushWrite(w)
}

func (n *TcpConn) LocalAddr() net.Addr {
	return nil
}

func (n *TcpConn) RemoteAddr() net.Addr {
	return n.conn.RemoteAddr()
}

func (n *TcpConn) Close() {
	if atomic.CompareAndSwapUint32(&n.closeFlag, 0, 1) == true {
		n.doPushWrite(nil)
	}
}

func (n *TcpConn) Destroy() {
	if n.conn != nil {
		_ = n.conn.Close()
	}
}

func (n *TcpConn) Run() error {
	defer n.recover()

	go func() {
		defer n.recover()
		if n.Write() == false {
			n.Close()
			n.Destroy()
		}
	}()

	if n.Recv() == false {
		return errors.New("tcp transport recv error")
	}
	return nil
}

func (n *TcpConn) Pong(nowTick int64) {
	if n.IsClosed() == true {
		return
	}

	pingMgr.DoPong(n, nowTick)
}

func (n *TcpConn) DoPong(nowTick int64) {
	n.lastPingTick = nowTick

	if n.connType != Linker_TCP_InComming {
		return
	}

	n.DoPing()
}

func (n *TcpConn) Ping() bool {
	if n.IsClosed() == true {
		return false
	}

	duration := n.option.ping_time
	checkTime := UnixTS() - n.lastPingTick
	if checkTime <= (duration*2 + 500) {
		if n.connType == Linker_TCP_OutGoing {
			n.DoPing()
		}
		return true
	}

	atomic.StoreInt32(&n.pingClose, 1)
	n.Close()
	return false
}

func (n *TcpConn) DoPing() {
	wrapper := &TransportMsgPack{
		msgType: 'T',
		msgID:   0,
		Buf:     nil,
	}
	n.doPushWrite(wrapper)
}

func (n *TcpConn) recover() {
	r := recover()
	if r == nil {
		return
	}
	slog.LogError("tcp_recovery", "recover error :%v", r)
	slog.LogError("tcp_recovery", "%s", string(debug.Stack()))
	n.Close()
	n.Destroy()
}

func (n *TcpConn) GetOption() *TransportOption {
	return n.option
}
