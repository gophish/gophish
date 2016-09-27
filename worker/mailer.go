package worker

import (
	"gopkg.in/gomail.v2"
	"time"
)

type email struct {
	message   *gomail.Message
	onError   func(err error)
	onSuccess func()
}

func createPool(d *gomail.Dialer, size int, buffer int) chan *email {
	ch := make(chan *email, buffer)

	for i := 0; i < size; i++ {
		go func() {
			var s gomail.SendCloser
			var err error
			open := false
			for {
				select {
				case m, ok := <-ch:
					if !ok {
						if open {
							if err := s.Close(); err != nil {
								Logger.Println(" Unable to close SMTP connection ", err)
							}
						}
						return
					}
					if !open {
						if s, err = d.Dial(); err != nil {
							go m.onError(err)
						} else {
							open = true
						}
					}
					if open {
						if err := gomail.Send(s, m.message); err != nil {
							go m.onError(err)
						} else {
							go m.onSuccess()
						}
					}
				// Close the connection to the SMTP server if no email was sent in
				// the last 30 seconds.
				case <-time.After(30 * time.Second):
					if open {
						if err := s.Close(); err != nil {
							Logger.Println(" Unable to close SMTP connection ", err)
						}
						open = false
					}

				}
			}
		}()
	}

	return ch
}
