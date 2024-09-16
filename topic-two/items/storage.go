package items

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateItemTable() error
	CreateItem(*Item) error
	UpdateValue(int, float64) error
	UpdateItem(*Item) error
	GetAllItems() ([]*Item, error)
	GetItem(int) (*Item, error)
	DeleteItem(int) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=dev dbname=debezium_test password=eatsleepcode sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (store *PostgresStore) CreateItemTable() error {
	query := `CREATE TABLE IF NOT EXISTS item(
		id serial primary key,
		name varchar(50),
		value numeric
	)`

	_, err := store.db.Exec(query)
	return err
}

func (store *PostgresStore) DropItemTable() error {
	query := `DROP TABLE IF EXISTS item`

	_, err := store.db.Query(query)
	return err
}
