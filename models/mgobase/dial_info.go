package mgobase

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	mgo "gopkg.in/mgo.v2"
)

var mgoModes = map[string]mgo.Mode{
	"eventual":  mgo.Eventual,
	"monotonic": mgo.Monotonic,
	"mono":      mgo.Monotonic,
	"strong":    mgo.Strong,
}

// DialInfo is the wrapper of mgo.DialInfo.
type DialInfo struct {
	*mgo.DialInfo
	Mode        mgo.Mode
	Refresh     bool
	SyncTimeout time.Duration
}

func extractURL(s string) (options map[string]string, mgoURL string, err error) {
	u, err := url.Parse(s)
	if err != nil {
		return
	}

	options = make(map[string]string)
	query := u.Query()
	for _, key := range []string{
		"dial_timeout",
		"sync_timeout",
		"mode",
	} {
		if val := query.Get(key); val != "" {
			options[key] = val
		}
		query.Del(key)
	}

	u.RawQuery = query.Encode()
	mgoURL = u.String()
	return
}

// ParseDialInfo parse the mongo dsn url to `DialInfo`.
func ParseDialInfo(dbURL string) (*DialInfo, error) {
	options, mgoURL, err := extractURL(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url(%s) with error: %s", dbURL, err)
	}

	info, err := mgo.ParseURL(mgoURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url(%s) with error: %s", dbURL, err)
	}
	dialInfo := &DialInfo{
		DialInfo:    info,
		Mode:        mgo.Mode(-1),
		SyncTimeout: -1,
	}

	for key, val := range options {
		switch key {
		case "dial_timeout":
			sec, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("parse timeout of database url(%s) with error: %s", dbURL, err)
			}
			dialInfo.Timeout = time.Second * time.Duration(sec)

		case "sync_timeout":
			sec, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("parse timeout of database url(%s) with error: %s", dbURL, err)
			}
			dialInfo.SyncTimeout = time.Second * time.Duration(sec)

		case "mode":
			mode, ok := mgoModes[val]
			if !ok {
				return nil, fmt.Errorf("invalid mode of database url(%s)", val)
			}
			dialInfo.Mode = mode
			dialInfo.Refresh = true

		}
	}

	return dialInfo, nil
}
