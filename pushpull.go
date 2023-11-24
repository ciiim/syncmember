package syncmember

func (s *SyncMember) pushPull() {
	for {
		select {
		case <-s.pushPullTicker.C:
			s.logger.Debug("pushPullTicker")
			s.doPushPull()
		case <-s.stopCh:
			s.logger.Debug("pushPullTicker stop")
			return
		}
	}
}

func (s *SyncMember) doPushPull() {
	target := kRamdonNodes(s.config.PushPullNums, s.nodes, func(n *Node) bool {
		return !n.IsCredible()
	})
	if len(target) == 0 {
		s.logger.Debug("no nodes can pushPull")
		return
	}

}
