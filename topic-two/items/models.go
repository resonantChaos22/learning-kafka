package items

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
)

type Item struct {
	ID    int
	Name  string
	Value float64
}

func (store *PostgresStore) CreateItem(item *Item) error {
	query := `INSERT INTO items (name, value) VALUES ($1, $2)`

	_, err := store.db.Exec(context.Background(), query, item.Name, item.Value)
	return err
}

func (store *PostgresStore) UpdateValue(id int, value float64) error {
	query := `UPDATE items SET value=$1 WHERE id=$2`

	roundedValue := math.Round(value*100) / 100
	_, err := store.db.Exec(context.Background(), query, roundedValue, id)
	return err
}

func (store *PostgresStore) UpdateItem(item *Item) error {
	query := `UPDATE items SET name=$1, value=$2 WHERE id=$3`

	_, err := store.db.Exec(context.Background(), query, item.Name, item.Value, item.ID)
	return err
}

func (store *PostgresStore) GetAllItems() ([]*Item, error) {
	query := `SELECT * FROM items`
	rows, err := store.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*Item{}
	for rows.Next() {
		item, err := scanIntoItem(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (store *PostgresStore) GetItem(id int) (*Item, error) {
	query := `SELECT * FROM items WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	store.LogPoolStats()
	rows, err := store.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		return scanIntoItem(rows)
	}

	return nil, fmt.Errorf("unable to find item with id %d", id)
}

func (store *PostgresStore) DeleteItem(id int) error {
	query := `DELETE FROM items WHERE id=$1`
	_, err := store.db.Exec(context.Background(), query, id)

	return err
}

func scanIntoItem(rows pgx.Rows) (*Item, error) {
	item := new(Item)
	err := rows.Scan(&item.ID, &item.Name, &item.Value)
	if err != nil {
		log.Println("ERROR")
		return nil, err
	}

	return item, nil
}

func (store *PostgresStore) LogPoolStats() {
	stats := store.db.Stat()
	log.Printf("Connections In Use: %d, Connections Available: %d, Total Connections: %d",
		stats.AcquiredConns(), stats.IdleConns(), stats.TotalConns())
}
