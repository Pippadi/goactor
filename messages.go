package actor

func SendStopMsg(destination Inbox) {
	destination <- func(a Actor) error {
		a.stop(nil)
		return nil
	}
}

func SendErrorMsg(destination Inbox, err error) {
	destination <- func(a Actor) error {
		return a.HandleError(err)
	}
}

func sendFarewell(destination Inbox, myInbox Inbox, reason error) {
	destination <- func(a Actor) error {
		return unregisterNested(a, myInbox, reason)
	}
}
