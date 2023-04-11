package actor

type Inbox chan Message

type Message func(a Actor) error

type Actor interface {
	Inbox() Inbox
	CreatorInbox() Inbox
	SpawnNested(Actor, string) (Inbox, error)
	Initialize() error
	IsStopping() bool
	Finalize()
	HandleError(error) error
	HandleLastMsg(Actor, error) error
	ID() string

	initialize(Inbox, string, bool) Inbox
	finalize()

	stop(error)
	okToStop() bool

	registerNested(Actor)
	stopAllNested()
	unregisterNested(Inbox) Actor
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

func initialize(a Actor, creatorInbox Inbox, id string, asRoot bool) (Inbox, error) {
	ibox := a.initialize(creatorInbox, id, asRoot)
	err := a.Initialize()
	if err != nil {
		return nil, err
	}
	return ibox, nil
}

func finalize(a Actor) {
	a.Finalize()
	a.finalize()
}

func launch(a Actor, creatorInbox Inbox, id string, asRoot bool) (Inbox, error) {
	ibox, err := initialize(a, creatorInbox, id, asRoot)
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

func SpawnRoot(a Actor, id string) (creatorInbox, nestedInbox Inbox, err error) {
	creatorInbox = make(Inbox, 1) // Only meant to be closed
	nestedInbox, err = launch(a, creatorInbox, id, true)
	return
}
