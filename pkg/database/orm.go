package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type Model interface {
	TableName() string
}

type ORM struct {
	db *sql.DB
}

func NewORM(db *sql.DB) *ORM {
	return &ORM{db: db}
}

func (o *ORM) Create(model Model) error {
	val := reflect.ValueOf(model).Elem()
	typ := val.Type()

	fields := []string{}
	placeholders := []string{}
	values := []interface{}{}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get("db") != "" && !val.Field(i).IsZero() {
			fields = append(fields, field.Tag.Get("db"))
			placeholders = append(placeholders, "?")
			values = append(values, val.Field(i).Interface())
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		model.TableName(),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))

	_, err := o.db.Exec(query, values...)
	return err
}

func (o *ORM) Find(model Model, id int) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", model.TableName())
	row := o.db.QueryRow(query, id)

	val := reflect.ValueOf(model).Elem()
	fields := make([]interface{}, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		fields[i] = val.Field(i).Addr().Interface()
	}

	return row.Scan(fields...)
}

// Add Update, Delete, Where methods...
