package main

import "database/sql"

type ResponseBuilder struct {
	Serialiser *Serialiser
}

func NewResoponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		Serialiser: NewSerialiser(),
	}
}

func (b *ResponseBuilder) BuildIndexResponse(rows *sql.Rows, fields map[string]*Field) (map[string]interface{}, error) {
	data, err := b.Serialiser.SerialiseColl(rows, fields)

	if err != nil {
		return map[string]interface{}{}, err
	}

	response := map[string]interface{}{
		"response": map[string]interface{}{
			"records": data,
		},
	}

	return response, nil
}

func (b *ResponseBuilder) BuildShowResponse(rows *sql.Rows, fields map[string]*Field) (map[string]interface{}, int, error) {
	data, err := b.Serialiser.SerialiseColl(rows, fields)

	if err != nil {
		return map[string]interface{}{}, 0, err
	}

	rowsCount := len(data)

	if rowsCount == 0 {
		return map[string]interface{}{}, 0, nil
	}

	response := map[string]interface{}{
		"response": map[string]interface{}{
			"record": data[0],
		},
	}

	return response, len(data), nil
}

func (b *ResponseBuilder) BuildCreateResponse(id int64, primaryKey string) map[string]interface{} {
	response := map[string]interface{}{
		"response": map[string]interface{}{
			primaryKey: id,
		},
	}
	return response
}

func (b *ResponseBuilder) BuildUpdateResponse(rowsAffected int64) map[string]interface{} {
	response := map[string]interface{}{
		"response": map[string]interface{}{
			"updated": rowsAffected,
		},
	}
	return response
}

func (b *ResponseBuilder) BuildDeleteResponse(rowsAffected int64) map[string]interface{} {
	response := map[string]interface{}{
		"response": map[string]interface{}{
			"deleted": rowsAffected,
		},
	}
	return response
}
