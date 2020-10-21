package log

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/evalphobia/logrus_fluent"
	"github.com/sirupsen/logrus"
)

const fluentPort = 24224

// NewFluentHook return a fluent hook for logrus logging
func NewFluentHook(level logrus.Level, endpoint string, tag string) error {
	// parse url
	url, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("Can't parse fluent endpoint: %v", err)
	}

	// parse port, use default one if none provided
	port := fluentPort
	if url.Port() != "" {
		port, err = strconv.Atoi(url.Port())
		if err != nil {
			return err
		}
	}

	hook, err := logrus_fluent.NewWithConfig(logrus_fluent.Config{
		Host:      url.Hostname(),
		Port:      port,
		Timeout:   1 * time.Second,
		MaxRetry:  10,
		RetryWait: 100,
	})
	if err != nil {
		return err
	}

	// set loglevel to level defined in config
	hook.SetLevels([]logrus.Level{level})
	hook.SetTag(tag)

	AddHook(hook)
	WithoutContext().Info("Fluent hook setup")

	return err
}
