package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
)

type Pet struct {
	ID   string
	Type string
}

type User struct {
	ID   string
	Name string

	Pets []*Pet
}

func Test_ScanFromDB(t *testing.T) {
	db, err := sql.Open("postgres",
		"postgresql://postgres:password@localhost:5432/postgres?sslmode=disable")
	checkErr(err)
	defer db.Close()
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFunc()
	checkErr(db.PingContext(timeout))

	rows, err := db.Query(`
		SELECT u.id,
		       u.name,
		       p.type AS pet_type,
		       p.id   AS pet_id
		FROM		 users u
		       JOIN pet p ON u.id = p.user_id
	`)
	checkErr(err)

	u := make([]*User, 0)

	scanRows(&rowsi{Rows: rows}, &u, prepareResultMap())

	litter.Dump(u)
}

type testRows struct {
	columns []string
	rows    [][]interface{}
	current int
}

func (r *testRows) Next() bool {
	r.current++
	return r.current <= len(r.rows)
}

func (r *testRows) Columns() ([]string, error) {
	return r.columns, nil
}

func (r *testRows) Scan(dst ...interface{}) error {
	for i := 0; i < len(dst); i++ {
		dst[i] = r.rows[r.current-1][i]
	}
	return nil
}

func Test_ScanFromMemory(t *testing.T) {
	u := make([]*User, 0)

	row1 := []interface{}{stringPtr("user1"), stringPtr("First user"), stringPtr("cat"), stringPtr("pet1")}
	row2 := []interface{}{stringPtr("user2"), stringPtr("Second user"), stringPtr("mouse"), stringPtr("pet3")}
	row3 := []interface{}{stringPtr("user1"), stringPtr("First user"), stringPtr("dog"), stringPtr("pet2")}

	rows := &testRows{
		columns: []string{"id", "name", "pet_type", "pet_id"},
		rows:    [][]interface{}{row1, row2, row3},
	}

	scanRows(rows, &u, prepareResultMap())

	litter.Dump(u)

	assert.Len(t, u, 2)
	assert.Len(t, u[0].Pets, 2)
	assert.Equal(t, "user1", u[0].ID)
	assert.Equal(t, "First user", u[0].Name)
	assert.Equal(t, "pet1", u[0].Pets[0].ID)
	assert.Equal(t, "cat", u[0].Pets[0].Type)
	assert.Equal(t, "pet2", u[0].Pets[1].ID)
	assert.Equal(t, "dog", u[0].Pets[1].Type)
}

func stringPtr(s string) *string {
	return &s
}

func prepareResultMap() *ResultMap {
	m := &ResultMap{
		PKDBToStruct: []string{"id"},
		DBToStruct:   make(map[string]string),
		Sub:          make(map[string]*ResultMap),
	}

	m.DBToStruct["id"] = "ID"
	m.DBToStruct["name"] = "Name"

	p := ResultMap{
		PKDBToStruct: []string{"id"},
		DBToStruct:   make(map[string]string),
		Sub:          make(map[string]*ResultMap),
	}
	p.DBToStruct["id"] = "ID"
	p.DBToStruct["type"] = "Type"

	m.Sub["pet"] = &p
	return m
}

/*
CREATE TABLE public.users
(
  id   VARCHAR PRIMARY KEY NOT NULL,
  name VARCHAR             NOT NULL
);

CREATE TABLE public.pet
(
  id      VARCHAR PRIMARY KEY NOT NULL,
  type    VARCHAR             NOT NULL,
  user_id VARCHAR             NOT NULL,
  CONSTRAINT pet_users_id_fk FOREIGN KEY (user_id) REFERENCES public.users (id)
);

INSERT INTO users (id, name) VALUES ('user1', 'First user');
INSERT INTO pet (id, type, user_id) VALUES ('pet1', 'cat', 'user1');
INSERT INTO pet (id, type, user_id) VALUES ('pet2', 'dog', 'user1');

*/
