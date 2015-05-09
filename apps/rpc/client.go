package rpc

type client struct {
}

func (this *client) Dial(addr string) {
}

func (this *client) Call(method string, args interface{}, reply interface{}) (err error) {
	return
}

func (this *client) Survey(method string, args interface{}, reply interface{}) (err error) {
	return
}

func (this *client) Go(method string, args interface{}, reply interface{}, done chan *Call) {
}

func (this *client) Close() error {
	return nil
}

func NewClient() *client {
	c := &client{}
	return c
}
