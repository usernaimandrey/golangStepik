package main

import (
	"database/sql"
)

type Serialiser struct {
	Values  []interface{}
	ScanArg []interface{}
}

func NewSerialiser() *Serialiser {
	return &Serialiser{}
}

func (s *Serialiser) SerialiseColl(rows *sql.Rows, fields map[string]*Field) ([]map[string]interface{}, error) {
	data := []map[string]interface{}{}
	coloumns, err := rows.Columns()

	if err != nil {
		return data, err
	}

	var scanError error

	for rows.Next() {
		s.prepareData(coloumns)
		err := rows.Scan(s.ScanArg...)

		if err != nil {
			scanError = err
			break
		}

		row, err := s.row(coloumns, fields)

		if err != nil {
			scanError = err
			break
		}
		data = append(data, row)
	}
	return data, scanError
}

func (s *Serialiser) prepareData(columns []string) {
	columnCount := len(columns)

	s.Values = make([]interface{}, columnCount)
	s.ScanArg = make([]interface{}, columnCount)
	for i := range s.Values {
		s.ScanArg[i] = &s.Values[i]
	}
}

func (s *Serialiser) row(columns []string, fields map[string]*Field) (map[string]interface{}, error) {
	row := map[string]interface{}{}

	var err error
	for i, v := range columns {
		value := s.Values[i]
		var valueConvert interface{}
		columnType := fields[v].Type

		if value != nil {
			valueConvert, err = DataTypeConverter(string(s.Values[i].([]byte)), columnType)
			if err != nil {
				break
			}

		} else {
			valueConvert = nil
		}
		row[v] = valueConvert
	}

	return row, err
}
