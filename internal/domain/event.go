package domain

type EventKind string

const (
	EventKindDefend EventKind = "defend"
	EventKindAttack EventKind = "attack"
	EventKindWar    EventKind = "war"
)

type EventTransition string

const (
	EventTransitionStarted   EventTransition = "started"
	EventTransitionSucceeded EventTransition = "succeeded"
	EventTransitionFailed    EventTransition = "failed"
)

type OngoingEvent struct {
	ID   int
	Kind EventKind
}

type WarEvent struct {
	Season int
}

type EventMessage struct {
	Kind        EventKind
	Transition  EventTransition
	DefendEvent *DefendEvent
	AttackEvent *AttackEvent
	WarEvent    *WarEvent
}
