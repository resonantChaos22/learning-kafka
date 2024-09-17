package items

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
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
	db *pgxpool.Pool
}

func NewPostgresStore() (*PostgresStore, error) {
	// connStr := "user=dev dbname=debezium_test password=eatsleepcode sslmode=disable"
	connStr := "postgres://dev:eatsleepcode@localhost:5432/debezium_test"
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	// config.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
	// 	color.Green("Connection to Postgres Established!")
	// 	return nil
	// }
	config.MaxConns = 25

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: pool,
	}, nil
}

func (store *PostgresStore) CreateItemTable() error {
	query := `CREATE TABLE IF NOT EXISTS items(
		id serial primary key,
		name varchar(50),
		value float
	)`

	_, err := store.db.Exec(context.Background(), query)
	return err
}

func (store *PostgresStore) DropItemTable() error {
	query := `DROP TABLE IF EXISTS items`

	_, err := store.db.Query(context.Background(), query)
	return err
}
