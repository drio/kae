package main

import (
	"time"
)

func (s *Server) runBackgroundJob(sleepSecs time.Duration) {
	for {
		listTokens, err := s.model.GetTokens()
		if err != nil {
			s.logger.Printf("runBackgroundJob: error getting tokens: %s", err)
			return
		}

		for _, t := range listTokens {
			lastHB, err := s.model.LastHeartBeat(t.ID)
			if err != nil {
				s.logger.Printf("runBackgroundJob: error getting last heartbeat: %s", err)
				return
			}

			secsSincelastHB := time.Now().Unix() - lastHB.Unix()
			hbInValidRange := secsSincelastHB <= int64(t.Interval)

			if t.Fired && !hbInValidRange {
				continue
			}

			var fireValue bool
			if t.Fired && hbInValidRange {
				s.logger.Printf("runBackgroundJob: clearing for token id:%d", t.ID)
				fireValue = false
			}

			if !t.Fired && !hbInValidRange {
				s.logger.Printf("runBackgroundJob: firing for token id:%d", t.ID)
				fireValue = true
			}

			err = s.model.Fire(t.ID, fireValue)
			if err != nil {
				s.logger.Printf("runBackgroundJob: error setting fire for tokenID=%d err=%s", t.ID, err)
				return
			}
		}

		s.logger.Printf("runBackgrondJob: Sleeping for %d secs", sleepSecs)
		time.Sleep(sleepSecs * time.Second)
	}
}
