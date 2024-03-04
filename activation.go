package main

// ActivationHandler type
type ActivationHandler struct {
}

func startActivation(actChannel chan *ActivationHandler, quit chan bool) {
	a := &ActivationHandler{}

	// put the reference to our struct in the channel
	// then continue to the loop
	actChannel <- a

forever:
	for {
		select {
		case <-quit:
			break forever
		}
	}
	logger.Debugf("Activation goroutine exiting")
	quit <- true
}

// Query activation state
func (a ActivationHandler) query() bool {
	return lengActive
}
