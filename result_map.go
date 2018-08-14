package main

import "strings"

type ResultMap struct {
	PKDBToStruct []string
	DBToStruct   map[string]string

	Sub map[string]*ResultMap
}

func NewDBValues(cols []string, vals []interface{}) *DBValues {
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

func (v *DBValues) filterByPrefix(prefix string) *DBValues {
	colMap := make(map[string]interface{})

	for key, value := range v.m {
		if strings.HasPrefix(key, prefix) {
			colMap[strings.TrimPrefix(key, prefix+"_")] = value
		}
	}

	return &DBValues{m: colMap}
}
