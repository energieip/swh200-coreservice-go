package database

const (
	//RETHINKDB database
	RETHINKDB = "rethinkdb"
)

type databaseError struct {
	s string
}

func (e *databaseError) Error() string {
	return e.s
}

// NewError raise an error
func NewError(text string) error {
	return &databaseError{text}
}

// DatabaseInterface database abstraction layer
type DatabaseInterface interface {
	Initialize(config DatabaseConfig) error
	CreateDB(dbName string) error
	CreateTable(dbName, tableName string, model interface{}) error
	InsertRecord(dbName, tableName string, data interface{}) (string, error)
	UpdateRecord(dbName, tableName, id string, data interface{}) error
	GetRecords(dbName, tableName string, criteria interface{}) ([]interface{}, error)
	GetRecord(dbName, tableName string, criteria interface{}) (interface{}, error)
	FetchAllRecords(dbName, tableName string) ([]interface{}, error)
	DeleteRecord(dbName, tableName string, data interface{}) error
	ListenTableChange(dbName, tableName string) (*DBCursor, error)
	ListenDBChange(dbName string) (*DBCursor, error)
	ListenFilterTableChange(dbName, tableName string, criteria interface{}) (*DBCursor, error)
	Close() error
}

//DatabaseConfig configuration structure
type DatabaseConfig struct {
	IP               string
	Port             string
	User             string //for authentification
	Password         string
	ServerCertificat string
	ClientCertificat string
	ClientKey        string
}

// NewNetwork instanciate the appropriate networkinterface
func NewDatabase(protocol string) (DatabaseInterface, error) {
	switch protocol {
	case RETHINKDB:
		return &RethinkbDatabase{}, nil
	default:
		return nil, NewError("Unknow databse type " + protocol)
	}
}
