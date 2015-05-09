package nano

func init() {
	endpointPool.byid = make(map[EndpointId]*pipeEndpoint)
	endpointPool.nextidChan = make(chan EndpointId, defaultChanLen)

	go endpointIdGenerator()
	go messagePoolWatchdog()
}
