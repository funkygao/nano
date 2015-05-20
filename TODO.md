TODO
====

- [ ] batch policy
  can be configured to accumulate no more than a fixed number of messages and to wait no longer than some fixed latency bound
  (say 100 messages or 5 seconds)
- [X] sync.Pool and bytes.Buffer
- [X] optional handshake
- [ ] benchmark shows mem leakage
- [ ] prefork in protocol implementation
- [ ] logger of Socket/ProtocolSocket
- [ ] multiple frames in one message
- [ ] socket identity for persistent socket
- [X] batch the framed msg to increase throughput
  protocol will have to manually call ep.Flush
- [ ] device use sendfile for zero copy
- [ ] newPipeEndpoint will be recycled in pool
- [ ] tls demo
- [ ] high performance timer, timewheel
- [X] snappy
- [ ] props/options key is int instead of string
- [X] bufio
- [ ] big size msg
- [X] demo Device usage
  examples/pipeline
- [ ] msg copy between transports
- [ ] Asynchronous DNS queries

### MQ

- compact stale log
- sendfile
- timeout

Our topic is divided into a set of totally ordered partitions, each of which is consumed by one consumer at any given time. 

A message is considered "committed" when all in sync replicas for that partition have applied it to their log. Only committed messages are ever given out to the consumer. 
