nano
====

MOM based

It frames messages.

Actually, behind the scenes, it does a lot. It frames messages. It sends and receives them in an asynchronous non-blocking manner. It checks for connection failures and re-establishes connections as needed. It queues the messages if the peer is unavailable at the moment. It ensures that individual peers are assigned their fair share of server resources so that a single client can't hijack the server. It routes the replies to the original requester.

intermediary nodes

通讯模式，它改变了通讯都基于一对一的连接这个假设。1:1 => N:M

通信模型

ip is hop-to-hop, udp/tcp is end-to-end

实现路由功能的组件叫作 Device


Pluggable Transports and Protocols

nanomsg implements priorities for outbound traffic. You may decide that messages are to be routed to a particular destination in preference, and fall back to an alternative destination only if the primary one is not available.

When connecting, you can optionally specify the local interface to use for the connection, like this: nn_connect (s, "tcp://eth0;192.168.0.111:5555").

Asynchronous DNS queries


ZeroCopy, what if mmap+sendfile

Partial failure is handled by the protocol, not by the user. In fact, it is transparent to the user.



tcp://interface:port 


    sock, err := protocol.xxx.NewSocket()
    sock.AddTransport(transport.yy.NewTransport())
    sock.SetOption(k, v)

    client:
        err := sock.Dial()
        sock.SendMsg(msg)

    server:
        err := sock.Listen()
        msg, err := sock.RecvMsg()


### Term

- Endpoint
  Message based, can be thought of as one side of tcp/ipc/etc

- Protocol
  
- Pipe
  goroutine safe full-duplex message-oriented connection between 2 peers

- Transport
  Responsible to provide concrete PipeDialer and PipeListener


### SP RFC

#### Connection initiation

The protocol header is 8 bytes long


        0                   1                   2                   3
        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |      0x00     |      0x53/S   |      0x50/P   |    version    |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |             type              |           reserved            |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

   BigEndian


    client                                  server
      |                                       |
      |  header(magic+req)                    |
      |-------------------------------------->|
      |                                       |
      |                 header(magic+rep)     |
      |<--------------------------------------|
      |                                       |
      |                                       |

#### Message

      uint64           []byte
    +------------+-----------------+
    | size (64b) |     payload     |
    +------------+-----------------+
