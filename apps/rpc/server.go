package rpc

type server struct {
}

func (this *server) Register() {
}

func (this *server) Serv(addr string) {
}

func NewServer() *server {
	s := &server{}
	return s
}
