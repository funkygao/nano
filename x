Pipe => connPipe
Dialer => dialer
Listener=> listener

Port, Endpoint => pipeEndpoint

PipeListener => tranport creates it
PipeDialer -> transport creates it

transport -> NewConnPipe





socket implements ProtocolSocket and Socket both.



PipeDialer, PipeListener => handshake
Pipe                     => stream IO


protocol.NewSocket -> nano.MakeSocket(proto) -> protocol.Init



sock.SendMsg -> sock.sendChan -> protocol.senderGoRoutine -> protocol.ep.SendMsg -> connPipe.SendMsg



sock.Dial  
    go dialer.dialer()
    go proto.sender()
    go proto.recver()
