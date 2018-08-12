package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/sanity-io/litter"
	"log"
	"reflect"
)

type Pet struct {
	ID   string
	Name string
}

type User struct {
	ID   string
	Name string

	Pets []Pet
}

type ResultMap struct {
	PKDBToStruct []string
	DBToStruct   map[string]string

	Sub map[string]*ResultMap
}

func NewColValues(cols []string, vals []interface{}) *DBValues {
	colMap := make(map[string]interface{})
	for i, col := range cols {
		colMap[col] = vals[i]
	}

	return &DBValues{m: colMap}
}

type DBValues struct {
	m map[string]interface{}
}

func (v *DBValues) getByName(s string) interface{} {
	for key, value := range v.m {
		if key == s {
			return value
		}
	}
	return nil
}

func scanRow(cols []string, vals []interface{}, resultSlicePtr interface{}, m *ResultMap) {
	colValues := NewColValues(cols, vals)
	foundInterface := findExistedByPK(resultSlicePtr, colValues, m)
	var found reflect.Value
	if foundInterface == nil {
		found = reflect.New(reflect.TypeOf(resultSlicePtr).Elem().Elem().Elem())
		fillStruct(found, colValues, m)
		s := reflect.ValueOf(resultSlicePtr).Elem()
		s.Set(reflect.Append(s, found))
	} else {
		found = reflect.ValueOf(foundInterface)
	}
}

func fillStruct(structValuePtr reflect.Value, values *DBValues, m *ResultMap) {
	structValue := structValuePtr.Elem()
	for dbColName, structFieldName := range m.DBToStruct {
		colValuePtr := values.m[dbColName] // got *string
		colValue := reflect.ValueOf(colValuePtr).Elem()
		structValue.FieldByName(structFieldName).Set(colValue)
	}
}

func findExistedByPK(resultSlicePtr interface{}, colValues *DBValues, m *ResultMap) interface{} {
	resultSliceValue := reflect.ValueOf(resultSlicePtr).Elem()

	for i := 0; i < resultSliceValue.Len(); i++ {
		if isEqualByPK(colValues, resultSliceValue.Index(i), m) {
			return resultSliceValue.Index(i).Interface()
		}
	}
	return nil
}

func isEqualByPK(dbValue *DBValues, itemOfSlice reflect.Value, resultMap *ResultMap) bool {
	for _, pkDBColName := range resultMap.PKDBToStruct {
		structValue := itemOfSlice.FieldByName(resultMap.DBToStruct[pkDBColName])
		colValue := dbValue.getByName(pkDBColName)
		if structValue.Interface() != colValue {
			return false
		}
	}
	return true
}

func main() {
	db, err := sql.Open("postgres",
		"postgresql://postgres:@db:5432/postgres?sslmode=disable")
	checkErr(err)
	defer db.Close()

	checkErr(db.Ping())

	m := &ResultMap{
		PKDBToStruct: make([]string, 0),
		DBToStruct:   make(map[string]string),
		Sub:          make(map[string]*ResultMap),
	}

	m.DBToStruct["id"] = "ID"
	m.DBToStruct["name"] = "Name"

	p := ResultMap{
		DBToStruct: make(map[string]string),
		Sub:        make(map[string]*ResultMap),
	}
	p.DBToStruct["name"] = "Name"

	m.Sub["pet"] = &p

	rows, err := db.Query(`
SELECT u.id,
       u.name,
       p.name AS pet_name,
       p.id   AS pet_id
FROM users u
       JOIN pet p ON u.id = p.user_id
`)
	checkErr(err)

	u := make([]*User, 0)

	for rows.Next() {
		cols, err := rows.Columns()
		checkErr(err)

		rawResult := make([]string, len(cols))
		//result := make([]string, len(cols))

		dest := make([]interface{}, len(cols)) // A temporary interface{} slice
		for i := range rawResult {
			dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
		}

		err = rows.Scan(dest...)

		scanRow(cols, dest, &u, m)

		checkErr(err)
	}

	litter.Dump(u)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
