package models

type Mailbox struct {
    Name string
}

func (m *Mailbox) ToRune() []rune {
    return []rune(m.Name)
}
