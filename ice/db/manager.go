package db

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/dyweb/gommon/errors"
	dlog "github.com/dyweb/gommon/log"

	"github.com/at15/go.ice/ice/config"
)

// backlog
// - each service can register which table it is using in manager, so it can print out the relationship between services

type Manager struct {
	mu       sync.RWMutex
	config   config.DatabaseManagerConfig
	log      *dlog.Logger
	wrappers map[string]*Wrapper
}

func NewManager(config config.DatabaseManagerConfig) *Manager {
	m := &Manager{
		config:   config,
		wrappers: make(map[string]*Wrapper, 1),
	}
	m.log = dlog.NewStructLogger(log, m)
	return m
}

func (mgr *Manager) DefaultName() (string, error) {
	if mgr.config.Default == "" {
		return "", errors.New("default database is not specified")
	}
	return mgr.config.Default, nil
}

func (mgr *Manager) Default() (*Wrapper, error) {
	if name, err := mgr.DefaultName(); err != nil {
		return nil, err
	} else {
		return mgr.Wrapper(name, true)
	}
}

func (mgr *Manager) Wrapper(name string, useDatabase bool) (*Wrapper, error) {
	mgr.mu.Lock()
	if w, ok := mgr.wrappers[name]; ok {
		mgr.mu.Unlock()
		return w, nil
	}
	defer mgr.mu.Unlock()
	var (
		a   Adapter
		c   config.DatabaseConfig
		db  *sql.DB
		dsn string
		err error
	)
	if c, err = mgr.Config(name); err != nil {
		return nil, errors.Wrap(err, "can't get config for "+name)
	}
	adapterName := c.Adapter
	if a, err = GetAdapter(adapterName); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can't get %s adapter for database %s", adapterName, name))
	}
	if !useDatabase {
		c.DBName = ""
		log.Debug("set dbname to empty string, use this when you want to create database")
	}
	if dsn, err = a.FormatDSN(c); err != nil {
		return nil, errors.Wrapf(err, "can't use %s adapter to format dsn for database %s", adapterName, name)
	}
	mgr.log.Debugf("connect using dsn %s", dsn)
	// NOTE: sql.Open does not make connection, so it won't throw error if remote db server is not ready
	if db, err = sql.Open(a.DriverName(), dsn); err != nil {
		return nil, errors.Wrap(err, "can't open database handle")
	}
	w := NewWrapper(a)
	w.SetDB(db)
	mgr.wrappers[name] = w
	return w, nil
}

func (mgr *Manager) Close() error {
	// TODO: error group
	var lastErr error
	for name, w := range mgr.wrappers {
		if err := w.Close(); err != nil {
			lastErr = errors.Wrapf(err, "error closing %s", name)
			log.Warnf("error closing %s %v", name, err)
		} else {
			log.Debugf("closed %s", name)
		}
	}
	return lastErr
}

func (mgr *Manager) Config(name string) (config.DatabaseConfig, error) {
	var known []string
	for _, d := range mgr.config.Databases {
		if d.Name == name {
			return d, nil
		}
		known = append(known, d.Name)
	}
	return config.EmptyDatabaseConfig, errors.Errorf("%s is not in known configs %s", name, known)
}

func (mgr *Manager) DefaultConfig() (config.DatabaseConfig, error) {
	var cfg config.DatabaseConfig
	name, err := mgr.DefaultName()
	if err != nil {
		return cfg, err
	}
	return mgr.Config(name)
}

func (mgr *Manager) PrintConfig() {
	if mgr == nil {
		log.Warn("Manager is nil pointer")
		return
	}
	fmt.Printf("default %s\n", mgr.config.Default)
	for i, c := range mgr.config.Databases {
		fmt.Printf("%d %s\n", i, c.String())
	}
}

func (mgr *Manager) RegisteredDrivers() []string {
	return Drivers()
}

func (mgr *Manager) RegisteredAdapters() []string {
	return Adapters()
}
