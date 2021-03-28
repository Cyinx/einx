package network

type TransportOption struct {
	msg_max_length uint32
	msg_max_count  int32 //max msg count per seconds
	ping_time      int64
	enable_ping    bool
}

func newTransportOption() TransportOption {
	o := TransportOption{
		msg_max_length: MSG_MAX_BODY_LENGTH,
		msg_max_count:  MSG_DEFAULT_COUNT,
		ping_time:      5 * 1000,
		enable_ping:    true,
	}
	return o
}

type OptionMgr interface {
	GetOption() *TransportOption
}

type Option = func(...interface{})

func Name(name string) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.name = name
		case *TcpClientMgr:
			v.name = name
		default:
			panic("option network name unknown type")
		}
	}
}

func Module(m EventReceiver) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.module = m
		case *TcpClientMgr:
			v.module = m
		default:
			panic("option network module unknown type")
		}
	}
}

func ListenAddr(addr string) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.addr = addr
		default:
			panic("option network listen addr unknown type")
		}
	}
}

func ServeHandler(serve_handler SessionHandler) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.agentHandler = serve_handler
		case *TcpClientMgr:
			v.agent_handler = serve_handler
		default:
			panic("option network serve handler unknown type")
		}
	}
}

func TransportMaxCount(c int) Option {
	return func(args ...interface{}) {
		if t, ok := args[0].(OptionMgr); ok == true {
			t.GetOption().msg_max_count = int32(c)
		} else {
			panic("option network transport keep alive unknown type")
		}
	}
}

func TransportMaxLength(c int) Option {
	return func(args ...interface{}) {
		if t, ok := args[0].(OptionMgr); ok == true {
			t.GetOption().msg_max_length = uint32(c)
		} else {
			panic("option network transport keep alive unknown type")
		}
	}
}


func TransportKeepAlive(e bool, c int64) Option {
	return func(args ...interface{}) {
		if t, ok := args[0].(OptionMgr); ok == true {
			t.GetOption().ping_time = c
			t.GetOption().enable_ping = e
		} else {
			panic("option network transport keep alive unknown type")
		}
	}
}
