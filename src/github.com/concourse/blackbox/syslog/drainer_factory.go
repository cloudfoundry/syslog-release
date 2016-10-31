package syslog

type DrainerFactory interface {
	NewDrainer() (Drainer, error)
}

type drainerFactory struct {
	destination Drain
	hostname    string
}

func NewDrainerFactory(destination Drain, hostname string) DrainerFactory {
	return &drainerFactory{
		destination: destination,
		hostname:    hostname,
	}
}

func (f *drainerFactory) NewDrainer() (Drainer, error) {
	return NewDrainer(
		f.destination,
		f.hostname,
	)
}
