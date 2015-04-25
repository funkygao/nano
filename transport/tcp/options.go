package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

// options is used for shared GetOption/SetOption logic.
type options map[string]interface{}

// GetOption retrieves an option value.
func (o options) get(name string) (interface{}, error) {
	if v, ok := o[name]; !ok {
		return nil, nano.ErrBadOption
	} else {
		return v, nil
	}
}

// SetOption sets an option.
func (o options) set(name string, val interface{}) error {
	switch name {
	case nano.OptionNoDelay:
		fallthrough
	case nano.OptionKeepAlive:
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

func newOptions() options {
	o := make(map[string]interface{})
	o[nano.OptionNoDelay] = true
	o[nano.OptionKeepAlive] = true
	return options(o)
}

func (o options) configTCP(conn *net.TCPConn) error {
	if v, ok := o[nano.OptionNoDelay]; ok {
		if err := conn.SetNoDelay(v.(bool)); err != nil {
			return err
		}
	}
	if v, ok := o[nano.OptionKeepAlive]; ok {
		if err := conn.SetKeepAlive(v.(bool)); err != nil {
			return err
		}
	}
	return nil
}
