package syslog

import (
	"errors"
	"time"

	sl "github.com/papertrail/remote_syslog2/syslog"
)

type Drain struct {
	Transport string `yaml:"transport"`
	Address   string `yaml:"address"`
}

//go:generate counterfeiter . Drainer

type Drainer interface {
	Drain(line string, tag string) error
}

const ServerPollingInterval = 5 * time.Second

type drainer struct {
	logger   *sl.Logger
	hostname string
}

func NewDrainer(drain Drain, hostname string) (*drainer, error) {
	err := errors.New("non-nil")
	var logger *sl.Logger

	for err != nil {
		logger, err = sl.Dial(
			hostname,
			drain.Transport,
			drain.Address,
			nil,
			30*time.Second,
			30*time.Second,
			9990,
		)

		if err != nil {
			time.Sleep(ServerPollingInterval)
		}
	}

	if err != nil {
		return nil, err
	}

	return &drainer{
		logger:   logger,
		hostname: hostname,
	}, nil
}

func (d *drainer) Drain(line string, tag string) error {
	d.logger.Packets <- sl.Packet{
		Severity: sl.SevInfo,
		Facility: sl.LogUser,
		Hostname: d.hostname,
		Tag:      tag,
		Time:     time.Now(),
		Message:  line,
	}

	select {
	case err := <-d.logger.Errors:
		return err
	default:
		return nil
	}
}
