package plugin

import "github.com/sirupsen/logrus"

type (
	Ext struct {
		Debug bool
	}
	Plugin struct {
		Ext Ext
	}
)

func (p Plugin) Exec() error {
	logrus.Debug("debug log\n")
	logrus.Info("info log\n")
	logrus.Warn("warn log\n")
	return nil
}
