package actor

import "log"

var _ Actor = new(Base)

type Base struct {
	inbox       Inbox
	parentInbox Inbox
	stopping    bool
	stopReason  error
	children    map[Inbox]Actor
	amRoot      bool
	id          string
}

func (b *Base) Inbox() Inbox       { return b.inbox }
func (b *Base) ParentInbox() Inbox { return b.parentInbox }

func (b *Base) initialize(parentInbox Inbox, id string, amRoot bool) Inbox {
	b.amRoot = amRoot
	b.inbox = make(Inbox, 16)
	b.parentInbox = parentInbox
	b.children = make(map[Inbox]Actor)
	b.stopping = false
	b.stopReason = nil
	b.id = id
	return b.Inbox()
}

func (b *Base) finalize() {
	if !b.amRoot {
		sendDisownMeMsg(b.parentInbox, b.Inbox(), b.stopReason)
	} else {
		if b.stopReason != nil {
			log.Println("Stopped because", b.stopReason)
		}
		close(b.ParentInbox())
	}
	close(b.Inbox())
}

func (b *Base) Initialize() error { return nil }
func (b *Base) Finalize()         {}

func (b *Base) ID() string       { return b.id }
func (b *Base) IsStopping() bool { return b.stopping }

func (b *Base) HandleError(err error) error {
	if b.amRoot {
		b.stop(err)
	} else {
		SendReportErrorMsg(b.parentInbox, err)
	}
	return nil
}

func (b *Base) LaunchAsChild(a Actor, id string) (Inbox, error) {
	ibox, err := launch(a, b.Inbox(), id, false)
	b.adoptChild(a)
	return ibox, err
}

func (b *Base) stop(reason error) {
	b.stopReason = reason
	b.stopping = true
	b.stopAllChildren()
}

func (b *Base) okToStop() bool {
	return b.stopping && len(b.children) == 0
}

func (b *Base) adoptChild(a Actor) {
	b.children[a.Inbox()] = a
}

func disownChild(parent Actor, childInbox Inbox, reason error) error {
	return parent.HandleDisown(parent.removeChild(childInbox), reason)
}
func (b *Base) removeChild(childInbox Inbox) Actor {
	child := b.children[childInbox]
	delete(b.children, childInbox)
	return child
}
func (b *Base) HandleDisown(child Actor, reason error) error { return reason }

func (b *Base) stopAllChildren() {
	for ibox, _ := range b.children {
		b.stopChild(ibox)
	}
}

func (b *Base) stopChild(childInbox Inbox) {
	SendStopMsg(childInbox)
}
