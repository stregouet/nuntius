package imap

import (
    "github.com/emersion/go-imap"
    "github.com/emersion/go-imap/client"
)

func toSeqSet(uids []uint32) *imap.SeqSet {
	var set imap.SeqSet
	for _, uid := range uids {
		set.AddNum(uid)
	}
	return &set
}

func fetch(c *client.Client, uids []uint32, items []imap.FetchItem, cb func(*imap.Message) error) error {
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

	set := toSeqSet(uids)
	if err := c.UidFetch(set, items, messages); err != nil {
		return err
	}
	if err := <-done; err != nil {
		return err
	}
    return nil
}
