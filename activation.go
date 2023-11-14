package main

// ToggleData type
type ToggleData struct {
	Mode uint
	Data uint
}

// ActivationHandler type
type ActivationHandler struct {
	queryChannel  chan bool
	toggleChannel chan ToggleData
	setChannel    chan bool
}

func startActivation(actChannel chan *ActivationHandler, quit chan bool) {
	a := &ActivationHandler{}

	a.queryChannel = make(chan bool)
	a.toggleChannel = make(chan ToggleData)
	a.setChannel = make(chan bool)

	// put the reference to our struct in the channel
	// then continue to the loop
	actChannel <- a

forever:
	for {
		select {
		case <-quit:
			break forever
		case <-a.queryChannel:
			a.queryChannel <- lengActive
		case v := <-a.toggleChannel:
			if v.Mode == 1 {
				lengActive = !lengActive
			} else {
				lengActive = false
			}
			a.queryChannel <- lengActive
		case v := <-a.setChannel:
			lengActive = v
			a.setChannel <- lengActive
		}
	}
	logger.Debugf("Activation goroutine exiting")
	quit <- true
}

// Query activation state
func (a ActivationHandler) query() bool {
	a.queryChannel <- true
	return <-a.queryChannel
}

// Set activation state
func (a ActivationHandler) set(v bool) bool {
	a.setChannel <- v
	return <-a.setChannel
}

// Toggle activation state on or off
func (a ActivationHandler) toggle(reactivationDelay uint) bool {
	data := ToggleData{
		Mode: 1,
		Data: reactivationDelay,
	}
	a.toggleChannel <- data
	return <-a.queryChannel
}

// Like toggle(), but only from on to off. Toggling when off will restart the
// timer.
func (a ActivationHandler) toggleOff(timeout uint) bool {
	a.toggleChannel <- ToggleData{
		Mode: 2,
		Data: timeout,
	}
	return <-a.queryChannel
}
