package network

type TransportOption struct {
	msg_max_length uint16
	msg_max_count  int32 //max msg count per seconds
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
			v.agent_handler = serve_handler
		case *TcpClientMgr:
			v.agent_handler = serve_handler
		default:
			panic("option network serve handler unknown type")
		}
	}
}

func TransportMaxCount(c int) Option {
	return func(args ...interface{}) {
		t := args[0]
		switch v := t.(type) {
		case *TcpServerMgr:
			v.option.msg_max_count = int32(c)
		case *TcpClientMgr:
			v.option.msg_max_count = int32(c)
		default:
			panic("option network transport max count unknown type")
		}
	}
}
