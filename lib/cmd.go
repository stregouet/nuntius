package lib

import (
	"errors"
	"strings"
)

const PARTIAL_CHAR = '>'

var UnfinishedValueErr = errors.New("unfinished quoted value")

type CmdArgs map[string]string

type Command struct {
	Partial bool
	Name    string
	Args    CmdArgs
}

type cmdParser struct {
	s string
}

func (p *cmdParser) len() int {
	return len(p.s)
}
func (p *cmdParser) empty() bool {
	return len(p.s) == 0
}
func (p *cmdParser) peek() byte {
	return p.s[0]
}
func (p *cmdParser) skipSpace() {
	p.s = strings.TrimLeft(p.s, " \t")
}

func (p *cmdParser) consume(b byte) bool {
	if p.empty() || p.peek() != b {
		return false
	}
	p.s = p.s[1:]
	return true
}
func (p *cmdParser) getWord(delim byte) string {
	p.skipSpace()
	if p.empty() {
		return ""
	}
	i := 0
	for {
		if i >= len(p.s) {
			break
		}
		if p.s[i] == delim {
			break
		}
		i++
	}
	var word string
	word, p.s = p.s[:i], p.s[i:]
	return word
}

func (p *cmdParser) getValue() (string, error) {
	if p.consume('"') {
		word := p.getWord('"')
		if !p.consume('"') {
			return "", UnfinishedValueErr
		}
		return word, nil
	} else {
		return p.getWord(' '), nil
	}
}

func ParseCmd(userinput string) (*Command, error) {
	p := cmdParser{userinput}
	partial := p.consume(PARTIAL_CHAR)
	name := p.getWord(' ')
	res := &Command{Partial: partial, Name: name, Args: make(CmdArgs)}
	for {
		key := p.getWord(':')
		if key == "" {
			break
		}
		p.consume(':')
		value, err := p.getValue()
		if err != nil {
			return nil, err
		}
		res.Args[key] = value
	}

	return res, nil
}
