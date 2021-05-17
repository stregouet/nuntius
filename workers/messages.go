package workers

type Message interface {
    GetId() int
    SetId(i int)
}

func WithId(m Message, id int) Message {
    m.SetId(id)
    return m
}

type BaseMessage struct {
    id int
}

func (b *BaseMessage) SetId(i int) {
    b.id = i
}

func (b *BaseMessage) GetId() int {
    return b.id
}

type Error struct {
    BaseMessage
    Error error
}

type FetchMailboxRes struct {
    BaseMessage
    List []string
}
type FetchMailbox struct {
    BaseMessage
    Mailbox string
}

