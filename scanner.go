package main

import (
	"github.com/sanity-io/litter"
	"reflect"
)

func scanRows(rows Rows, u interface{}, m *ResultMap) {
	cols, err := rows.Columns()
	checkErr(err)

	for rows.Next() {
		rawResult := make([]string, len(cols))

		dest := make([]interface{}, len(cols)) // A temporary interface{} slice
		for i := range rawResult {
			dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
		}

		err = rows.Scan(dest...)

		litter.Dump(dest)

		colValues := NewDBValues(cols, dest)
		scanRow(colValues, u, m)

		checkErr(err)
	}
}

func scanRow(colValues *DBValues, resultSlicePtr interface{}, m *ResultMap) {
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

	for prefix, subResultMap := range m.Sub {
		subColValues := colValues.filterByPrefix(prefix)

		sliceField, _ := found.Type().Elem().FieldByName("Pets")
		sliceFieldType := sliceField.Type // []Pet
		//sliceElemType  := sliceFieldType.Elem() // Pet

		fieldValue := found.Elem().FieldByName("Pets")
		isNil := fieldValue.IsNil()
		if isNil {
			slice := reflect.MakeSlice(sliceFieldType, 0, 0)
			fieldValue.Set(slice)
		}

		subFieldPtr := fieldValue.Addr().Interface()

		scanRow(subColValues, subFieldPtr, subResultMap)
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

func isEqualByPK(dbValue *DBValues, itemOfSlicePtr reflect.Value, resultMap *ResultMap) bool {
	for _, pkDBColName := range resultMap.PKDBToStruct {
		itemOfSlice := itemOfSlicePtr.Elem()
		structValue := itemOfSlice.FieldByName(resultMap.DBToStruct[pkDBColName])
		colValueInterfacePtr := dbValue.getByName(pkDBColName)
		colValueStringPtr := colValueInterfacePtr.(*string)

		if structValue.Interface() != *colValueStringPtr {
			return false
		}
	}
	return true
}
