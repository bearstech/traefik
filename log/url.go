package log

import (
	"log/syslog"
	"net/url"
	"strings"

	"github.com/evalphobia/logrus_fluent"
	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

func Parse(_url string) (logrus.Hook, error) {
	u, err := url.Parse(_url)
	if err != nil {
		return nil, err
	}
	hp := strings.Split(u.Host, ":")
	var host string, port int
	if len(hp) == 2 {
		host = hp[0]
		port, err = strconv.Atoi(hp[1])
		if err != nil {
			return nil, err
		}
	} else {
		host = hp
		port = -1
	}
	switch u.Scheme {
	case "syslog+udp":
		return lSyslog.NewSyslogHook("udp", u.Host, syslog.LOG_INFO, u.Path)
	case "syslog+tcp":
		return lSyslog.NewSyslogHook("tcp", u.Host, syslog.LOG_INFO, u.Path)
	case "syslog+unix":
		return lSyslog.NewSyslogHook("", u.Host, syslog.LOG_INFO, u.Path)
	case "fluentd":
		if port == -1 {
			port = 24224
		}
		return logrus_fluent.NewWithConfig(logrus_fluent.Config{
			Host: host,
			Port: port,
		})
	case: "graylog":
		return graylog.NewGraylogHook(u.Host), nil
	}

	return nil, nil
}
