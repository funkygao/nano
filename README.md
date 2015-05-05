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
- It ensures that individual peers are assigned their fair share of server resources so that a single client can't hijack the server. 
- It routes the replies to the original requester.

Enjoy!

### Design

#### Pluggable Transport

Currently supported transports:

- tcp

  `tcp://[eth0;]<host>:<port>`

- ipc

  `ipc://<unix_domain_socket_path>`

- tls

  `tls+tcp://<host>:<port>`

#### Pluggable Protocol

Currently supported protocols:

- bus
- pubsub
- pipeline
- pair
- reqrep
- survey

### Internals

#### Data Flow

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

