package models

type Message struct {
    Subject string
}

func (m *Message) ToRune() []rune {
    return []rune(m.Subject)
}
