package mgobase

import (
	"fmt"
	"sync"

	mgo "gopkg.in/mgo.v2"
)

// Database is the db you connect.
type Database struct {
	session *mgo.Session
	dbName  string
}

// NewDatabase returns a Database instance.
// You should initiate it by calling `InitWithURL` or `InitWithDialInfo`.
func NewDatabase() *Database {
	return &Database{}
}

// InitWithURL initiates db instance by given url.
// The additional options are `dial_timeout`(second), `sync_timeout`(second) and `mode`(strong|eventual|mono|monotonic).
//
// See `mgo.Dial` also.
func (d *Database) InitWithURL(url string) error {
	info, err := ParseDialInfo(url)
	if err != nil {
		return err
	}
	return d.InitWithDialInfo(info)
}

// InitWithDialInfo initiates db instance by given DialInfo.
func (d *Database) InitWithDialInfo(info *DialInfo) error {
	if info == nil {
		return fmt.Errorf("dial info can't be nil")
	}

	session, err := mgo.DialWithInfo(info.DialInfo)
	if err != nil {
		return fmt.Errorf("mgo.DialWithInfo(%#v) with error: %s", info.DialInfo, err)
	}

	if info.Mode > -1 {
		session.SetMode(info.Mode, info.Refresh)
	}

	if info.SyncTimeout > -1 {
		session.SetSyncTimeout(info.SyncTimeout)
	}

	d.session = session
	d.dbName = info.Database

	return nil
}

// C creates a Collection which connects to a specific mongo collection with optional indexes.
func (d *Database) C(name string, indexes ...Index) *Collection {
	return &Collection{
		sessionFactory:  d.Session,
		dbName:          d.dbName,
		colName:         name,
		indexes:         indexes,
		ensureIndexLock: &sync.Mutex{},
	}
}

// Copy copies a new db instance from this instance.
//
// See `mgo.Session.Copy` also.
func (d *Database) Copy() *Database {
	return &Database{
		session: d.session.Copy(),
		dbName:  d.dbName,
	}
}

// Clone clones a new db instance from this instance, but reuses the same session as the original database.
func (d *Database) Clone() *Database {
	return &Database{
		session: d.session.Clone(),
		dbName:  d.dbName,
	}
}

// Close closes the database connection.
func (d *Database) Close() {
	d.session.Close()
}

// Session returns a new session.
func (d *Database) Session() *mgo.Session {
	return d.session.Copy()
}

// Drop drops the database.
func (d *Database) Drop() error {
	return d.session.DB(d.dbName).DropDatabase()
}
