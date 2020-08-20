package database

import (
	"fmt"

	"github.com/izhaohu/migrate/source"
)

var (
	ErrLocked = fmt.Errorf("cannot acquire lock")
	ErrFailed = fmt.Errorf("execute sql failed")
	drivers   = make(map[string]Driver)
)

type Driver interface {
	Open(uri string) (Driver, error)
	Close() error
	Lock(source.Module) error
	Unlock(source.Module) error
	Exec(string) error
	Version(source.Module) (int, bool, error)
	SetVer(source.Module, int, bool) error
}

func Register(n string, d Driver) {
	drivers[n] = d
}

func Open(n string, uri string) (Driver, error) {
	r, ok := drivers[n]
	if !ok {
		return nil, fmt.Errorf("invalid driver")
	}

	return r.Open(uri)
}
