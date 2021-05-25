package workers

import "github.com/stregouet/nuntius/models"

type Message interface {
    GetId() int
    SetId(i int)
    GetAccName() string
    SetAccName(accname string)
}

func WithId(m Message, id int) Message {
    m.SetId(id)
    return m
}

type BaseMessage struct {
    id int
    accountname string
}

func (b *BaseMessage) SetId(i int) {
    b.id = i
}

func (b *BaseMessage) GetId() int {
    return b.id
}
func (b *BaseMessage) GetAccName() string {
    return b.accountname
}
func (b *BaseMessage) SetAccName(accname string) {
    b.accountname = accname
}

type MsgToDb struct {
    BaseMessage
    Wrapped Message
}

type Done struct {
    BaseMessage
}
type Error struct {
    BaseMessage
    Error error
}


type FetchMailboxRes struct {
    BaseMessage
    List []*models.Thread
}
type FetchMailboxImapRes struct {
    BaseMessage
    Mailbox string
    Mails []*models.Mail
}
type FetchMailbox struct {
    BaseMessage
    Mailbox string
}

type FetchMailboxesRes struct {
    BaseMessage
    Mailboxes []*models.Mailbox
}
type FetchMailboxesImapRes struct {
    BaseMessage
    Mailboxes []*models.Mailbox
}
type FetchMailboxes struct {
    BaseMessage
}



type ConnectImap struct {
    BaseMessage
}
