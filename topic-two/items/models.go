package items

import (
	"database/sql"
	"fmt"
)

type Item struct {
	ID    int
	Name  string
	Value float64
}

func (store *PostgresStore) CreateItem(item *Item) error {
	query := `INSERT INTO item (name, value) VALUES ($1, $2)`

	_, err := store.db.Query(query, item.Name, item.Value)
	return err
}

func (store *PostgresStore) UpdateValue(id int, value float64) error {
	query := `UPDATE item SET value=$1 WHERE id=$2`

	_, err := store.db.Query(query, value, id)
	return err
}

func (store *PostgresStore) UpdateItem(item *Item) error {
	query := `UPDATE item SET name=$1, value=$2 WHERE id=$3`

	_, err := store.db.Query(query, item.Name, item.Value, item.ID)
	return err
}

func (store *PostgresStore) GetAllItems() ([]*Item, error) {
	query := `SELECT * FROM item`
	rows, err := store.db.Query(query)
	if err != nil {
		return nil, err
	}

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
	query := `SELECT * FROM item WHERE id=$1`
	rows, err := store.db.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoItem(rows)
	}

	return nil, fmt.Errorf("unable to find item with id %d", id)
}

func (store *PostgresStore) DeleteItem(id int) error {
	query := `DELETE FROM item WHERE id=$1`
	_, err := store.db.Query(query, id)

	return err
}

func scanIntoItem(rows *sql.Rows) (*Item, error) {
	item := new(Item)
	err := rows.Scan(item.ID, item.Name, item.Value)
	if err != nil {
		return nil, err
	}

	return item, nil
}
