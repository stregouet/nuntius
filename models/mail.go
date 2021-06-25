package models

import (
    "fmt"
    "strconv"
    "strings"
    "time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
)

type BodyPath string


func (bp BodyPath) ToMessagePath() ([]int, error) {
	s := strings.TrimPrefix(string(bp), "/")
    if s == "" {
        return []int{}, nil
    }
	parts := strings.Split(s, "/")
	res := make([]int, len(parts))
	for idx, p := range parts {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		res[idx] = parsed
	}
	return res, nil
}

func BodyPathFromMessagePath(p []int) BodyPath {
	if len(p) == 0 {
		return "/"
	}
    buf := new(strings.Builder)
    for _, p := range p {
        buf.WriteByte('/')
        buf.WriteString(strconv.Itoa(p))
    }
    return BodyPath(buf.String())
}

type BodyPart struct {
	Path        BodyPath `json:"path"`
	MIMEType    string   `json:"mimetype"`
	MIMESubType string   `json:"mimesubtype"`
}

func (bp *BodyPart) Depth() int {
	if bp.Path == "/" {
		return 0
	}
	return strings.Count(string(bp.Path), "/")
}

func (bp *BodyPart) ToRune() []rune {
	return []rune(fmt.Sprintf("%s/%s (%s)", bp.MIMEType, bp.MIMESubType, bp.Path))
}

func bodyPartsFromImapParts(bs *imap.BodyStructure, parts []*BodyPart, path []int) []*BodyPart {
	result := append(parts, &BodyPart{
		Path: BodyPathFromMessagePath(path),
		MIMEType: bs.MIMEType,
		MIMESubType: bs.MIMESubType,
	})
	if bs.Parts != nil {
		for i, part := range bs.Parts {
			result = bodyPartsFromImapParts(part, result, append(path, i))
		}
	}
	return result
}

func BodyPartsFromImap(bs *imap.BodyStructure) []*BodyPart {
	result := make([]*BodyPart, 0)
	return bodyPartsFromImapParts(bs, result, []int{})
}

type Mail struct {
	Id        int
	Uid       uint32
	Threadid  int
	Subject   string
	Flags     []string
	MessageId string
	Mailbox   string
	InReplyTo string
	Parts     []*BodyPart
	depth     int
	Date      time.Time
	Header    *mail.Header
}

func (m *Mail) ToRune() []rune {
	return []rune(fmt.Sprintf("%s %s",
		m.Date.Format("2006-01-02 15:04:05"),
		m.Subject,
	))
}

func (m *Mail) Depth() int {
	return m.depth
}

func (m *Mail) FindPlaintext() *BodyPath {
	if m.Parts == nil || len(m.Parts) == 0 {
		return nil
	}
	for _, p := range m.Parts {
		if p.MIMEType == "text" && p.MIMESubType == "plain" {
			return &p.Path
		}
	}
	return nil
}

func (m *Mail) FindFirstNonMultipart() *BodyPath {
	if m.Parts == nil || len(m.Parts) == 0 {
		return nil
	}
	for _, p := range m.Parts {
		if p.MIMEType != "multipart" {
			return &p.Path
		}
	}
	return nil
}
