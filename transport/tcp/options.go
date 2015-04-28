package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

type options map[string]interface{}

func newOptions() options {
	opt := make(options)
	opt[nano.OptionNoDelay] = true
	return opt
}

func (o options) get(name string) (interface{}, error) {
	if v, ok := o[name]; !ok {
		return nil, nano.ErrBadOption
	} else {
		return v, nil
	}
}

func (o options) set(name string, val interface{}) error {
	switch name {
	case nano.OptionNoDelay:
		switch v := val.(type) {
		case bool:
			o[name] = v
			return nil
		default:
			return nano.ErrBadValue
		}
	}
	return nano.ErrBadOption
}

func (o options) configTCP(conn *net.TCPConn) error {
	if v, ok := o[nano.OptionNoDelay]; ok {
		if err := conn.SetNoDelay(v.(bool)); err != nil {
			return err
		}
	}

	return nil
}
