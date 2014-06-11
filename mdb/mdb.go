package mdb

import (
	"labix.org/v2/mgo"
)

var (
	session   *mgo.Session
	db        *mgo.Database
	sserver   string
	sdb       string
	susername string
	spassword string
)

func SESS() *mgo.Session {
	if session == nil {
		connect()
	} else if len(session.LiveServers()) < 1 {
		connect()
	}
	return session
}

func DB() *mgo.Database {
	if db == nil {
		connect()
	} else if len(session.LiveServers()) < 1 {
		connect()
	}
	return db
}

func Init(server, db, username, password string) error {
	sserver = server
	sdb = db
	susername = username
	spassword = password
	//
	return connect()
}

func Close() {
	if session == nil {
		return
	}
	db = nil
	if len(session.LiveServers()) < 1 {
		session = nil
		return
	}
	session.Close()
	session = nil
}

func connect() error {
	var err error
	session, err = mgo.Dial(sserver)
	if err != nil {
		return err
	}
	db = session.DB(sdb)
	if len(susername) > 0 {
		err = db.Login(susername, spassword)
		if err != nil {
			return err
		}
	}
	return nil
}
