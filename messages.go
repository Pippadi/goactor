package actor

func SendStopMsg(destination Inbox) {
	destination <- func(a Actor) error {
		a.stop(nil)
		return nil
	}
}

func SendReportErrorMsg(destination Inbox, err error) {
	destination <- func(a Actor) error {
		return a.HandleError(err)
	}
}

func sendDisownMeMsg(destination Inbox, myInbox Inbox, reason error) {
	destination <- func(a Actor) error {
		return disownChild(a, myInbox, reason)
	}
}
