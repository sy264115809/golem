package mgobase

import (
	"sync"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	// Collection represents a mongo collection of db.
	Collection struct {
		sessionFactory  func() *mgo.Session
		dbName          string
		colName         string
		indexes         []Index
		ensureIndexed   bool
		ensureIndexLock *sync.Mutex

		debug  bool
		logger logger
	}

	// Index represents a index of collection.
	Index struct {
		Key        []string // Index key fields; prefix name with dash (-) for descending order
		Unique     bool     // Prevent two documents from having the same index key
		Background bool     // Build index in background and return immediately
		Sparse     bool     // Only index documents containing the Key fields

		// If ExpireAfter is defined the server will periodically delete
		// documents with indexed time.Time older than the provided delta.
		ExpireAfter time.Duration
	}
)

func (c *Collection) ensureIndex() error {
	if !c.ensureIndexed {
		sess := c.sessionFactory()
		defer func() {
			sess.Close()
			c.ensureIndexed = true
		}()

		col := sess.DB(c.dbName).C(c.colName)
		c.ensureIndexLock.Lock()
		defer c.ensureIndexLock.Unlock()

		for _, index := range c.indexes {
			err := col.EnsureIndex(mgo.Index{
				Key:         index.Key,
				Unique:      index.Unique,
				Background:  index.Background,
				Sparse:      index.Sparse,
				ExpireAfter: index.ExpireAfter,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var (
	slowQueryTime = 2 * time.Second
)

// SetSlowQueryTime sets the max second that a query can be tolerated before it's considered as a slow query.
func SetSlowQueryTime(sec uint) {
	slowQueryTime = time.Duration(sec) * time.Second
}

func logSlowQuery(from, to time.Time) {
	if slowQueryTime > 0 && globalLogger != nil {
		if queryDuration := to.Sub(from); queryDuration > slowQueryTime {
			warnf("[mgo]slow query takes %s exceeds expected time %s", queryDuration, slowQueryTime)
		}
	}
}

// Invoke invokes a callback function with a session created from session factory.
func (c *Collection) Invoke(fn func(*mgo.Collection) error) error {
	err := c.ensureIndex()
	if err != nil {
		return err
	}

	sess := c.sessionFactory()
	defer sess.Close()

	col := sess.DB(c.dbName).C(c.colName)

	start := time.Now()
	err = fn(col)
	time.Now().Sub(start)

	return parseMgoError(err)
}

// Insert inserts one or more documents.
func (c *Collection) Insert(models ...interface{}) error {
	return c.Invoke(func(col *mgo.Collection) error {
		return col.Insert(models...)
	})
}

// UpsertByObjectID upserts documents by given object id.
func (c *Collection) UpsertByObjectID(id bson.ObjectId, update interface{}) (info *mgo.ChangeInfo, err error) {
	if !bson.IsObjectIdHex(id.Hex()) {
		return nil, ErrInvalidID
	}

	err = c.Invoke(func(col *mgo.Collection) error {
		info, err = col.UpsertId(id, update)
		return err
	})

	return
}

// Upsert upserts documents by query selector.
func (c *Collection) Upsert(selector, update interface{}) (info *mgo.ChangeInfo, err error) {
	err = c.Invoke(func(col *mgo.Collection) error {
		info, err = col.Upsert(selector, update)
		return err
	})
	return
}

// UpdateByObjectID updates a single document by object id.
func (c *Collection) UpdateByObjectID(id bson.ObjectId, update interface{}) error {
	if !bson.IsObjectIdHex(id.Hex()) {
		return ErrInvalidID
	}
	return c.Invoke(func(col *mgo.Collection) error {
		return col.UpdateId(id, update)
	})
}

// Update updates a single document by query selector.
func (c *Collection) Update(selector, update interface{}) error {
	return c.Invoke(func(col *mgo.Collection) error {
		return col.Update(selector, update)
	})
}

// UpdateAll updates all documents match the query selector.
func (c *Collection) UpdateAll(selector, update interface{}) error {
	return c.Invoke(func(col *mgo.Collection) error {
		_, err := col.UpdateAll(selector, update)
		return err
	})
}

// UpdateSetByObjectID updates a single document with $set operator by object id.
func (c *Collection) UpdateSetByObjectID(id bson.ObjectId, update interface{}) error {
	if !bson.IsObjectIdHex(id.Hex()) {
		return ErrInvalidID
	}
	return c.Invoke(func(col *mgo.Collection) error {
		return col.UpdateId(id, bson.M{"$set": update})
	})
}

// UpdateSet updates a single document with $set operator by query selector.
func (c *Collection) UpdateSet(selector, update interface{}) error {
	return c.Invoke(func(col *mgo.Collection) error {
		return col.Update(selector, bson.M{"$set": update})
	})
}

// UpdateSetAll updates all documents match the query selector with $set operator.
func (c *Collection) UpdateSetAll(selector, update interface{}) error {
	return c.Invoke(func(col *mgo.Collection) error {
		_, err := col.UpdateAll(selector, bson.M{"$set": update})
		return err
	})
}

// Remove removes a single document by query selector.
func (c *Collection) Remove(selector interface{}) error {
	return c.Invoke(func(col *mgo.Collection) error {
		return col.Remove(selector)
	})
}

// RemoveByID removes a single document by object id in string form.
func (c *Collection) RemoveByID(id string) error {
	if !bson.IsObjectIdHex(id) {
		return ErrInvalidID
	}
	return c.RemoveByObjectID(bson.ObjectIdHex(id))
}

// RemoveByObjectID removes a single document by object id.
func (c *Collection) RemoveByObjectID(id bson.ObjectId) error {
	if !bson.IsObjectIdHex(id.Hex()) {
		return ErrInvalidID
	}
	return c.Invoke(func(col *mgo.Collection) error {
		return col.RemoveId(id)
	})
}

// RemoveAll removes all documents match the query selector.
func (c *Collection) RemoveAll(selector interface{}) error {
	return c.Invoke(func(col *mgo.Collection) error {
		_, err := col.RemoveAll(selector)
		return err
	})
}

// Find finds a single document by given query and sort conditions if exist.
func (c *Collection) Find(query, model interface{}, sorts ...string) error {
	return c.Invoke(func(col *mgo.Collection) error {
		return col.Find(query).Sort(sorts...).One(model)
	})
}

// FindByID finds a single document by object id in string form.
func (c *Collection) FindByID(id string, model interface{}) error {
	if !bson.IsObjectIdHex(id) {
		return ErrInvalidID
	}
	return c.FindByObjectID(bson.ObjectIdHex(id), model)
}

// FindByObjectID finds a single document by object id.
func (c *Collection) FindByObjectID(id bson.ObjectId, model interface{}) error {
	if !bson.IsObjectIdHex(id.Hex()) {
		return ErrInvalidID
	}
	return c.Invoke(func(col *mgo.Collection) error {
		return col.FindId(id).One(model)
	})
}

// FindAll finds all documents match the query, skips the `skip` steps and returns the limited sorted results with projection fields.
//
// `selector` is used for project fields if no nil.
//
// `limit` will be concerned if it's greater than 0.
//
// each elements of `sorts` should be nonempty string if the `sorts` are provided.
func (c *Collection) FindAll(query, selector, models interface{}, skip, limit int, sorts ...string) error {
	return c.Invoke(func(col *mgo.Collection) error {
		return col.Find(query).Select(selector).Skip(skip).Limit(limit).Sort(sorts...).All(models)
	})
}

// FindAllWithPagination works just like FindAll, but it returns a paginater to indicate the informations about pagination.
func (c *Collection) FindAllWithPagination(query, selector, models interface{}, skip, limit int, sorts ...string) (Paginater, error) {
	err := c.FindAll(query, selector, models, skip, limit, sorts...)
	if err != nil {
		return nil, err
	}

	count, err := c.Count(query)
	if err != nil {
		return nil, err
	}

	return NewPaginater(skip, limit, count), nil
}

// FindAllWithMarker uses the query-based paging technology. The pagination will use the `marker` to decide how to retrieve the documents.
//
// See `Marker` also.
func (c *Collection) FindAllWithMarker(query, selector, models interface{}, marker Marker, limit int) (prev, next interface{}, err error) {
	c.Invoke(func(col *mgo.Collection) error {
		prev, next, err = marker.List(col, query, selector, models, limit)
		return err
	})
	return
}

// Distinct unmarshals into result the list of distinct values for the given key.
func (c *Collection) Distinct(query, models interface{}, key string) error {
	return c.Invoke(func(col *mgo.Collection) error {
		return col.Find(query).Distinct(key, models)
	})
}

// Count returns the total number of documents by query.
func (c *Collection) Count(query interface{}) (n int, err error) {
	err = c.Invoke(func(col *mgo.Collection) error {
		n, err = col.Find(query).Count()
		return err
	})
	return
}

// Drop drops the collection.
func (c *Collection) Drop() (err error) {
	return c.Invoke(func(col *mgo.Collection) error {
		return col.DropCollection()
	})
}
