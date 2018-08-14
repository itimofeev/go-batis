package main

import (
	"database/sql"
	"reflect"
)

func scanRows(rows *sql.Rows, u interface{}, m *ResultMap) {
	for rows.Next() {
		cols, err := rows.Columns()
		checkErr(err)

		rawResult := make([]string, len(cols))

		dest := make([]interface{}, len(cols)) // A temporary interface{} slice
		for i := range rawResult {
			dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
		}

		err = rows.Scan(dest...)

		scanRow(cols, dest, u, m)

		checkErr(err)
	}
}

func scanRow(cols []string, vals []interface{}, resultSlicePtr interface{}, m *ResultMap) {
	colValues := NewDBValues(cols, vals)
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
