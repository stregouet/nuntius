package lib

// mix of xstate and https://venilnoronha.io/a-simple-state-machine-framework-in-go

type Transition struct {
	Target StateType
	Action Action
}

type Event struct {
	Transition TransitionType
	Payload    interface{}
}

type Transitions map[TransitionType]*Transition
type States map[StateType]*State

type Action func(machineContext interface{}, tr *Event)

type State struct {
	Entry       Action
	Exit        Action
	Transitions Transitions
}

type StateListener func(s *State, ev *Event)

type StateType string

func (s StateType) Matches(other StateType) bool {
	return s == other
}

type TransitionType string

type Machine struct {
	// the extended state
	Context interface{}
	States  States

	Current StateType

	listenerId          int
	transitionListeners map[int]StateListener
}

func NewMachine(ctx interface{}, initial StateType, states States) *Machine {
	return &Machine{
		Context: ctx,
		States: states,
		Current: initial,
		transitionListeners: make(map[int]StateListener),
	}
}

func (m *Machine) OnTransition(f StateListener) int {
	m.listenerId++
	m.transitionListeners[m.listenerId] = f
	return m.listenerId
}
func (m *Machine) OffTransition(listenerId int) {
	delete(m.transitionListeners, listenerId)
}

func (m *Machine) callListeners(current *State, ev *Event) {
	for _, l := range m.transitionListeners {
		l(current, ev)
	}
}

func (m *Machine) Send(ev *Event) {
	current := m.States[m.Current]
	if tr, ok := current.Transitions[ev.Transition]; ok {
		if current.Exit != nil {
			current.Exit(m.Context, ev)
		}
		m.Current = tr.Target
		nextState := m.States[tr.Target]
		if nextState.Entry != nil {
			nextState.Entry(m.Context, ev)
		}
		if tr.Action != nil {
			tr.Action(m.Context, ev)
		}
		m.callListeners(nextState, ev)
	}

}
