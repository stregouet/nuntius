package workers

import "github.com/stregouet/nuntius/models"

type Message interface {
	GetId() int
	SetId(i int)
	GetAccName() string
	SetAccName(accname string)
}

type ClonableMessage interface {
	Message
	Clone() Message
}

func WithId(m Message, id int) Message {
	m.SetId(id)
	return m
}

type BaseMessage struct {
	id          int
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

type Done struct {
	BaseMessage
}


type Error struct {
	BaseMessage
	Error error
}


type FetchThread struct {
	BaseMessage
	RootId       int
}

type FetchThreadRes struct {
	BaseMessage
	Mails        []*models.Mail
}

type FetchMailboxRes struct {
	BaseMessage
	List        []*models.Thread
	LastSeenUid uint32
}


type FetchNewMessages struct {
	BaseMessage
	Mailbox     string
	LastSeenUid uint32
}

type FetchNewMessagesRes struct {
	BaseMessage
	Mailbox string
	Mails   []*models.Mail
}

type InsertNewMessages struct {
	BaseMessage
	Mailbox string
	Mails   []*models.Mail
}

type InsertNewMessagesRes struct {
	BaseMessage
	Threads []*models.Thread
}

type FetchMailboxImapRes struct {
	BaseMessage
	Mailbox string
	Mails   []*models.Mail
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

func (m *FetchMailboxes) Clone() Message {
	return &FetchMailboxes{m.CloneBase()}
}

type ConnectImap struct {
	BaseMessage
}

