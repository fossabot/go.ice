// Code generated by gommon from db/gommon.yml DO NOT EDIT.
package db
import dlog "github.com/dyweb/gommon/log"

func (mgr *Manager) SetLogger(logger *dlog.Logger) {
	mgr.log = logger
}

func (mgr *Manager) GetLogger() *dlog.Logger {
	return mgr.log
}

func (mgr *Manager) LoggerIdentity(justCallMe func() *dlog.Identity) *dlog.Identity {
	return justCallMe()
}
