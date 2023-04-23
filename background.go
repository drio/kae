package main

import (
	"time"
)

type bgJobOpts struct {
	loop    bool
	delayFn func()
}

// Loop over the tokens and check the last heartbeat. Set the fire accordingly
// Otherwise, loop and run the sleep fun
func (s *Server) runBackgroundJob(opts bgJobOpts) {
	logic := func() {
		listTokens, err := s.model.GetTokens()
		if err != nil {
			s.logger.Printf("runBackgroundJob: error getting tokens: %s", err)
			return
		}

		for _, t := range listTokens {
			if t.Disabled {
				continue
			}

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

		opts.delayFn()
	}

	if opts.loop {
		for {
			logic()
		}
	}

	logic()
}
