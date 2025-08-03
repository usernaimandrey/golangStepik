package main

import "fmt"

type RecordNotFounError struct {
	Message string
}

type FieldInvalidTypeError struct {
	Message string
}

func (r *RecordNotFounError) Error() string {
	return r.Message
}

func (f *FieldInvalidTypeError) Error() string {
	return f.Message
}

func NewRecordNotFoundError() *RecordNotFounError {
	return &RecordNotFounError{Message: "record not found"}
}

func NewFieldInvalidTypeError(field string) *FieldInvalidTypeError {
	messge := fmt.Sprintf("field %s have invalid type", field)
	return &FieldInvalidTypeError{Message: messge}
}
