nano
====


Pluggable Transports and Protocols

nanomsg implements priorities for outbound traffic. You may decide that messages are to be routed to a particular destination in preference, and fall back to an alternative destination only if the primary one is not available.

When connecting, you can optionally specify the local interface to use for the connection, like this: nn_connect (s, "tcp://eth0;192.168.0.111:5555").

Asynchronous DNS queries


ZeroCopy, what if mmap+sendfile

Partial failure is handled by the protocol, not by the user. In fact, it is transparent to the user.



tcp://interface:port 


### Term

- Endpoint
  Message based, can be thought of as one side of tcp/ipc/etc

- Protocol
  
- Pipe
  goroutine safe full-duplex message-oriented connection between 2 peers


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
