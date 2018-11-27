package database

import (
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

//RethinkbDatabase implementation
type RethinkbDatabase struct {
	session *r.Session
}

type DBCursor = r.Cursor

//Initialize setup Database instance
func (d *RethinkbDatabase) Initialize(config DatabaseConfig) error {
	session, err := r.Connect(r.ConnectOpts{
		Address: config.IP,
	})
	if err != nil {
		return err
	}
	d.session = session
	return nil
}

//Close Database connexion
func (d *RethinkbDatabase) Close() error {
	if d.session != nil {
		return d.session.Close()
	}
	return nil
}

//CreateDB add new table in database
func (d *RethinkbDatabase) CreateDB(dbName string) error {
	if d.session == nil {
		return NewError("Database Error: session not connected")
	}
	_, err := r.DBCreate(dbName).RunWrite(d.session)
	return err
}

//CreateTable add new table in database
func (d *RethinkbDatabase) CreateTable(dbName, tableName string, model interface{}) error {
	if d.session == nil {
		return NewError("Database Error: session not connected")
	}
	_, err := r.DB(dbName).TableCreate(tableName).RunWrite(d.session)
	return err
}

//ListenTableChange listen table change
func (d *RethinkbDatabase) ListenTableChange(dbName, tableName string) (*DBCursor, error) {
	if d.session == nil {
		return nil, NewError("Database Error: session not connected")
	}
	return r.DB(dbName).Table(tableName).Changes().Run(d.session)
}

//ListenDBChange listen table change
func (d *RethinkbDatabase) ListenDBChange(dbName string) (*DBCursor, error) {
	if d.session == nil {
		return nil, NewError("Database Error: session not connected")
	}
	return r.DB(dbName).Changes().Run(d.session)
}

//ListenFilterTableChange listen table change
func (d *RethinkbDatabase) ListenFilterTableChange(dbName, tableName string, criteria interface{}) (*DBCursor, error) {
	if d.session == nil {
		return nil, NewError("Database Error: session not connected")
	}
	return r.DB(dbName).Table(tableName).Filter(criteria).Changes().Run(d.session)
}

//InsertRecord add a new record in database
func (d *RethinkbDatabase) InsertRecord(dbName, tableName string, data interface{}) (string, error) {
	if d.session == nil {
		return "", NewError("Database Error: session not connected")
	}
	result, err := r.DB(dbName).Table(tableName).Insert(data).RunWrite(d.session)
	if err != nil {
		return "", err
	}
	return result.GeneratedKeys[0], nil
}

//UpdateRecord add a record in database
func (d *RethinkbDatabase) UpdateRecord(dbName, tableName, id string, data interface{}) error {
	if d.session == nil {
		return NewError("Database Error: session not connected")
	}
	_, err := r.DB(dbName).Table(tableName).Get(id).Update(data).RunWrite(d.session)
	return err
}

//GetRecords return all matching record for criteria map
func (d *RethinkbDatabase) GetRecords(dbName, tableName string, criteria interface{}) ([]interface{}, error) {
	var records []interface{}
	if d.session == nil {
		return records, NewError("Database Error: session not connected")
	}
	rows, err := r.DB(dbName).Table(tableName).Filter(criteria).Run(d.session)
	if err != nil {
		return records, err
	}
	err = rows.All(&records)
	return records, err
}

//GetRecord return the first matching record for criteria map
func (d *RethinkbDatabase) GetRecord(dbName, tableName string, criteria interface{}) (interface{}, error) {
	var record interface{}
	if d.session == nil {
		return record, NewError("Database Error: session not connected")
	}
	cursor, err := r.DB(dbName).Table(tableName).Filter(criteria).Run(d.session)
	if err != nil {
		return nil, err
	}
	cursor.One(&record)
	cursor.Close()
	return record, nil
}

//FetchAllRecords get all database records
func (d *RethinkbDatabase) FetchAllRecords(dbName, tableName string) ([]interface{}, error) {
	var records []interface{}
	if d.session == nil {
		return records, NewError("Database Error: session not connected")
	}
	rows, err := r.DB(dbName).Table(tableName).Run(d.session)
	if err != nil {
		return records, err
	}
	err = rows.All(&records)
	return records, err
}

//DeleteRecord remove a record from the database
func (d *RethinkbDatabase) DeleteRecord(dbName, tableName string, data interface{}) error {
	if d.session == nil {
		return NewError("Database Error: session not connected")
	}
	_, err := r.DB(dbName).Table(tableName).Get(data).Delete().Run(d.session)
	return err
}
