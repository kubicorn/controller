package backoff

import "time"

type backoff struct {
	name string
	i    int
	n    int
}

func NewBackoff(name string) *backoff {
	return &backoff{
		name: name,
		i:    1,
		n:    1,
	}
}

func (b *backoff) Hang() {
	time.Sleep(time.Duration(b.i) * time.Second)
}
