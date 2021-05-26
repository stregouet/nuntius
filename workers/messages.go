package workers

import "github.com/stregouet/nuntius/models"

type Message interface {
    GetId() int
    SetId(i int)
    GetAccName() string
    SetAccName(accname string)
    Clone() Message
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
func (b *BaseMessage) CloneBase() BaseMessage {
    return BaseMessage{b.GetId(), b.GetAccName()}
}


type MsgToDb struct {
    BaseMessage
    Wrapped Message
}
func (m *MsgToDb) Clone() Message {
    return &MsgToDb{m.CloneBase(), m.Wrapped.Clone()}
}

type Done struct {
    BaseMessage
}
func (m *Done) Clone() Message {
    return &Done{m.CloneBase()}
}
type Error struct {
    BaseMessage
    Error error
}
func (m *Error) Clone() Message {
    return &Error{m.CloneBase(), m.Error}
}


type FetchMailboxRes struct {
    BaseMessage
    List []*models.Thread
}
func (m *FetchMailboxRes) Clone() Message {
    return &FetchMailboxRes{m.CloneBase(), m.List}
}
type FetchMailboxImapRes struct {
    BaseMessage
    Mailbox string
    Mails []*models.Mail
}
func (m *FetchMailboxImapRes) Clone() Message {
    return &FetchMailboxImapRes{m.CloneBase(), m.Mailbox, m.Mails}
}
type FetchMailbox struct {
    BaseMessage
    Mailbox string
}
func (m *FetchMailbox) Clone() Message {
    return &FetchMailbox{m.CloneBase(), m.Mailbox}
}

type FetchMailboxesRes struct {
    BaseMessage
    Mailboxes []*models.Mailbox
}
func (m *FetchMailboxesRes) Clone() Message {
    return &FetchMailboxesRes{m.CloneBase(), m.Mailboxes}
}
type FetchMailboxesImapRes struct {
    BaseMessage
    Mailboxes []*models.Mailbox
}
func (m *FetchMailboxesImapRes) Clone() Message {
    return &FetchMailboxesImapRes{m.CloneBase(), m.Mailboxes}
}
type FetchMailboxes struct {
    BaseMessage
}
func (m *FetchMailboxes) Clone() Message {
    return &FetchMailboxes{m.CloneBase()}
}



type ConnectImap struct {
    BaseMessage
}
func (m *ConnectImap) Clone() Message {
    return &ConnectImap{m.CloneBase()}
}
