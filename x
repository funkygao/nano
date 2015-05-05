
PipeDialer, PipeListener => handshake
Pipe                     => stream IO


protocol.NewSocket -> nano.MakeSocket(proto) -> protocol.Init



sock.SendMsg -> sock.sendChan -> protocol.senderGoRoutine -> protocol.ep.SendMsg -> connPipe.SendMsg



sock.Dial  
    go dialer.dialer()
    go proto.sender()
    go proto.recver()
