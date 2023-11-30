package syncmember

func (s *SyncMember) gossip() {
	for {
		select {
		case <-s.gossipTicker.C:
			s.doGossip()
		case <-s.stopCh:
			return
		}
	}
}

func (s *SyncMember) doGossip() {

}
