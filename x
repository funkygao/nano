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
