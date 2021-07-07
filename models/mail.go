package models

import (
    "fmt"
    "strconv"
    "strings"
    "time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/config"
	ndb "github.com/stregouet/nuntius/database"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/widgets"
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

func (bp *BodyPart) FindMatch(filters config.Filters) string {
	for mime, cmd := range filters {
		mimeparts := strings.Split(mime, "/")
		if len(mimeparts) != 2 {
			continue
		}
		if mimeparts[0] == bp.MIMEType {
			if mimeparts[1] == bp.MIMESubType || mimeparts[1] == "*" {
				return cmd
			}
		}
	}
	return ""
}

func (bp *BodyPart) Depth() int {
	if bp.Path == "/" {
		return 0
	}
	return strings.Count(string(bp.Path), "/")
}

func (bp *BodyPart) StyledContent() []*widgets.ContentWithStyle {
	return []*widgets.ContentWithStyle{
		widgets.NewContent(fmt.Sprintf("%s/%s (%s)", bp.MIMEType, bp.MIMESubType, bp.Path)),
	}
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

func (m *Mail) StyledContent() []*widgets.ContentWithStyle {
	s := tcell.StyleDefault.Bold(m.IsUnread())
	return []*widgets.ContentWithStyle{
		widgets.NewContent(m.Date.Format("2006-01-02 15:04:05") + " "),
		{m.Subject, s},
		widgets.NewContent(" (" + strings.Join(m.Flags, "|") + ")"),
	}
}

func (m *Mail) IsUnread() bool {
	for _, f := range m.Flags {
		if f == imap.SeenFlag {
			return false
		}
	}
	return true
}

func (m *Mail) Depth() int {
	return m.depth
}

func (m *Mail) FindPlaintext() *BodyPart {
	if m.Parts == nil || len(m.Parts) == 0 {
		return nil
	}
	for _, p := range m.Parts {
		if p.MIMEType == "text" && p.MIMESubType == "plain" {
			return p
		}
	}
	return nil
}

func (m *Mail) FindFirstNonMultipart() *BodyPart {
	if m.Parts == nil || len(m.Parts) == 0 {
		return nil
	}
	for _, p := range m.Parts {
		if p.MIMEType != "multipart" {
			return p
		}
	}
	return nil
}

func (m *Mail) Delete(r ndb.Execer) error {
	_, err := r.Exec("DELETE FROM mail WHERE id = ?", m.Id)
	return err
}

func (m *Mail) UpdateFlags(r ndb.Execer, flags []string) error {
	if lib.IsCountEqual(m.Flags, flags) {
		// newflags is same as already known flags => no need to perform update
		return nil
	}
	_, err := r.Exec("UPDATE mail SET flags = ? WHERE id = ?", strings.Join(flags, ","), m.Id)
	return err
}

func FetchMails(r ndb.Queryer, mailbox, accname string) ([]*Mail, error) {
	rows, err := r.Query(`SELECT m.id, m.uid FROM
        mail m
        JOIN mailbox mbox ON mbox.id = m.mailbox
        JOIN account a ON a.id = m.account AND a.id = mbox.account
      WHERE
        a.name = ? AND mbox.name = ?`, accname, mailbox)
	if err != nil {
		return nil, err
	}
	result := make([]*Mail, 0)
	for rows.Next() {
		var id int
		var uid int
		err = rows.Scan(&id, &uid)
		if err != nil {
			return nil, err
		}
		result = append(result, &Mail{Id: id, Uid: uint32(uid)})
	}
	return result, nil
}
