package actor

type Inbox chan Message

type Message func(a Actor) error

type Actor interface {
	Inbox() Inbox
	ParentInbox() Inbox
	LaunchAsChild(Actor, string) (Inbox, error)
	Initialize() error
	IsStopping() bool
	Finalize()
	HandleError(error) error
	HandleDisown(Actor, error) error
	ID() string

	initialize(Inbox, string, bool) Inbox
	finalize()

	stop(error)
	okToStop() bool

	adoptChild(Actor)
	stopChild(Inbox)
	stopAllChildren()
	removeChild(Inbox) Actor
}

func run(a Actor) {
	for !a.okToStop() {
		err := (<-a.Inbox())(a)
		if err != nil {
			errWhileHandling := a.HandleError(err)
			if errWhileHandling != nil {
				a.stop(errWhileHandling)
			}
		}
	}
}

func initialize(a Actor, parentInbox Inbox, id string, asRoot bool) (Inbox, error) {
	ibox := a.initialize(parentInbox, id, asRoot)
	return ibox, a.Initialize()
}

func finalize(a Actor) {
	a.Finalize()
	a.finalize()
}

func launch(a Actor, parentInbox Inbox, id string, asRoot bool) (Inbox, error) {
	ibox, err := initialize(a, parentInbox, id, asRoot)
	if err != nil {
		finalize(a)
		return nil, err
	}
	go func() {
		defer finalize(a)
		run(a)
	}()
	return ibox, nil
}

func LaunchAsRoot(a Actor, id string) (parentInbox, childInbox Inbox, err error) {
	parentInbox = make(Inbox, 1) // Only meant to be closed
	childInbox, err = launch(a, parentInbox, id, true)
	return
}
