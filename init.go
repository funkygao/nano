package nano

func init() {
	endpointPool.byid = make(map[uint32]*pipeEndpoint)
	endpointPool.nextidChan = make(chan uint32, defaultChanLen)

	go endpointIdGenerator()
	go messagePoolWatchdog()
}
