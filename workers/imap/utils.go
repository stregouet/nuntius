package imap

import (
	"crypto/tls"
	"fmt"
	"os/exec"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/pkg/errors"
)

func toSeqSet(uids []uint32) *imap.SeqSet {
	var set imap.SeqSet
	for _, uid := range uids {
		set.AddNum(uid)
	}
	return &set
}

func fetch(c *client.Client, set *imap.SeqSet, items []imap.FetchItem, cb func(*imap.Message) error) error {
	messages := make(chan *imap.Message)
	done := make(chan error)

	go func() {
		var err error
		for msg := range messages {
			err = cb(msg)
			if err != nil {
				// drain the channel upon error
				for range messages {
				}
			}
		}
		done <- err
	}()

	if err := c.UidFetch(set, items, messages); err != nil {
		return err
	}
	if err := <-done; err != nil {
		return err
	}
	return nil
}

func getPass(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", errors.Wrap(err, "cannot exec passcmd")
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func connectSmtps(serverName string, port uint16) (*smtp.Client, error) {
	host := fmt.Sprintf("%s:%d", serverName, port)
	conn, err := smtp.DialTLS(host, &tls.Config{
		ServerName: serverName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "while dialing tls for smtp")
	}
	return conn, nil
}

func newSaslClient(auth, user, password string) (sasl.Client, error) {
	var saslClient sasl.Client
	switch auth {
	case "":
		fallthrough
	case "none":
		saslClient = nil
	case "login":
		saslClient = sasl.NewLoginClient(user, password)
	case "plain":
		saslClient = sasl.NewPlainClient("", user, password)
	default:
		return nil, fmt.Errorf("Unsupported auth mechanism %s", auth)
	}
	return saslClient, nil
}

func listRecipients(h *mail.Header) ([]*mail.Address, error) {
	var rcpts []*mail.Address
	for _, key := range []string{"to", "cc", "bcc"} {
		list, err := h.AddressList(key)
		if err != nil {
			return nil, err
		}
		rcpts = append(rcpts, list...)
	}
	return rcpts, nil
}
