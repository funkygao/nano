nano
====

nano is an implementation in pure Go of the SP ("Scalable Protocols") protocols.

The reference implementation of the SP protocols is available as:
[nanomsg](http://www.nanomsg.org)

nano is easy to use, but behind the scenes, it does a lot:

- It frames messages. 
- It sends and receives messages in an asynchronous non-blocking manner. 
- It checks for connection failures and re-establishes connections as needed. 
- It queues the messages if the peer is unavailable at the moment. 
- It manages timeout and thread safety.
- It ensures that individual peers are assigned their fair share of server resources so that a single client can't hijack the server. 
- It allows you to send data to the topology rather than to particular endpoint.
- It compress/decompress IO streams on demand.

Enjoy!

### Design

#### Pluggable Transport

Nano is transport agnostic.

Currently supported transports:

- tcp

  `tcp://[eth0;]<host>:<port>`

- ipc

  `ipc://<unix_domain_socket_path>`

- tls

  `tls+tcp://<host>:<port>`

#### Pluggable Protocol

Nano is protocol agnostic.
Protocol implements the topology: a set of applications participating on the same aspect of the business logic.

Currently supported protocols:

- bus
- pubsub
- pipeline
- pair
- reqrep
- survey

### Internals

        Application, create msg                     Application, then msg.Free
            |                                           ^
            V                                           |
        Socket.SendMsg ---------------------------- Socket.RecvMsg
            |                                           |
        ProtocolSendHook.SendHook                   ProtocolRecvHook.RecvHook
            |                                           ^
            | ProtocolSocket.SendChannel                | ProtocolSocket.RecvChannel
            V                                           |
        Protocol                                    Protocol
            |                                           ^
            | sender thread loop                        | receiver thread loop
            V                                           |
        EndPoint.SendMsg                            EndPoint.RecvMsg
            |                                           |
        Pipe.SendMsg, then msg.Free                 Pipe.RecvMsg, create msg
            |                                           |
            |                                           |
            +-------------------------------------------+                              
                              transport 
           

#### Reducing GC Presure

#### Timers


### SP RFC

   BigEndian

#### Connection initiation

The protocol header is 8 bytes long


        0                   1                   2                   3
        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |      0x00     |      0x53/S   |      0x50/P   |    version    |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |             type              |           reserved            |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+


#### Message Frame

      uint64           []byte
    +------------+-----------------+
    | size (64b) |     payload     |
    +------------+-----------------+

