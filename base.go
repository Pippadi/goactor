package actor

import "log"

var _ Actor = new(Base)

type Base struct {
	inbox        Inbox
	creatorInbox Inbox
	stopping     bool
	stopReason   error
	nesteds      map[Inbox]Actor
	amRoot       bool
	id           string
}

func (b *Base) Inbox() Inbox        { return b.inbox }
func (b *Base) CreatorInbox() Inbox { return b.creatorInbox }

func (b *Base) initialize(creatorInbox Inbox, id string, amRoot bool) Inbox {
	b.amRoot = amRoot
	b.inbox = make(Inbox, 16)
	b.creatorInbox = creatorInbox
	b.nesteds = make(map[Inbox]Actor)
	b.stopping = false
	b.stopReason = nil
	b.id = id
	return b.Inbox()
}

func (b *Base) finalize() {
	if !b.amRoot {
		sendFarewell(b.creatorInbox, b.Inbox(), b.stopReason)
	} else {
		if b.stopReason != nil {
			log.Println("Stopped because", b.stopReason)
		}
		close(b.CreatorInbox())
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
		SendErrorMsg(b.creatorInbox, err)
	}
	return nil
}

func (b *Base) SpawnNested(a Actor, id string) (Inbox, error) {
	ibox, err := launch(a, b.Inbox(), id, false)
	b.registerNested(a)
	return ibox, err
}

func (b *Base) stop(reason error) {
	b.stopReason = reason
	b.stopping = true
	b.stopAllNested()
}

func (b *Base) okToStop() bool {
	return b.stopping && len(b.nesteds) == 0
}

func (b *Base) registerNested(a Actor) {
	b.nesteds[a.Inbox()] = a
}

func unregisterNested(creator Actor, nestedInbox Inbox, reason error) error {
	return creator.HandleLastMsg(creator.unregisterNested(nestedInbox), reason)
}
func (b *Base) unregisterNested(nestedInbox Inbox) Actor {
	nested := b.nesteds[nestedInbox]
	delete(b.nesteds, nestedInbox)
	return nested
}
func (b *Base) HandleLastMsg(nested Actor, reason error) error { return reason }

func (b *Base) stopAllNested() {
	for ibox, _ := range b.nesteds {
		b.stopNested(ibox)
	}
}

func (b *Base) stopNested(nestedInbox Inbox) {
	SendStopMsg(nestedInbox)
}
