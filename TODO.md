TODO
====

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
