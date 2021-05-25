package models

type Mailbox struct {
    Name string
    Count uint32
    Unseen uint32
    ReadOnly bool
}

func (m *Mailbox) ToRune() []rune {
    return []rune(m.Name)
}
