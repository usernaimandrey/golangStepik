package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	outValue, ok := out.(reflect.Value)

	if !ok {
		outValue = reflect.ValueOf(out)
	}

	if outValue.Kind() != reflect.Pointer {
		return fmt.Errorf("param out is not ptr")
	}

	dataValue := reflect.ValueOf(data)
	var err error

	if dataValue.Kind() == reflect.Map && outValue.Elem().Kind() == reflect.Struct {
		inDataMap := data.(map[string]interface{})
		err = MapStrategy(inDataMap, outValue)
	} else if dataValue.Kind() == reflect.Slice && outValue.Elem().Kind() == reflect.Slice {
		inDataSlice := data.([]interface{})
		t := outValue.Type().Elem().Elem()
		err = SliceStarategy(inDataSlice, outValue, t)
	} else {
		err = fmt.Errorf("unknown strategy")
	}

	return err
}

func MapStrategy(inDataMap map[string]interface{}, outValue reflect.Value) error {
	var err error
	outT := outValue.Elem().Type()
	for i := 0; i < outT.NumField(); i++ {
		field := outT.Field(i)
		fieldName := field.Name
		inputData := inDataMap[fieldName]

		fieldVal := outValue.Elem().FieldByName(fieldName)

		if !fieldVal.IsValid() {
			err = fmt.Errorf("field %s is not valid", fieldName)
			break
		}

		if !fieldVal.CanSet() {
			err = fmt.Errorf("field %s can not set", fieldName)
			break
		}

		if field.Type.Kind() == reflect.Struct {
			err = i2s(inputData, fieldVal.Addr())
			if err != nil {
				break
			}
			continue
		} else if field.Type.Kind() == reflect.Slice {
			err = i2s(inputData, fieldVal.Addr())
			if err != nil {
				break
			}
			continue
		}

		inputDataToSet := reflect.ValueOf(inputData)
		err = SetData(inputDataToSet, fieldVal, fieldName)
		if err != nil {
			break
		}
	}
	return err
}

func SliceStarategy(inDataSlice []interface{}, outValue reflect.Value, t reflect.Type) error {
	outSlice := outValue.Elem()
	var err error

	for _, v := range inDataSlice {
		ptr := reflect.New(t)

		mapValue := v.(map[string]interface{})
		err = i2s(mapValue, ptr)
		outSlice = reflect.Append(outSlice, ptr.Elem())
	}

	outValue.Elem().Set(outSlice)
	return err
}

func SetData(inputDataToSet, fieldVal reflect.Value, fieldName string) error {
	if inputDataToSet.Type().AssignableTo(fieldVal.Type()) {
		fieldVal.Set(inputDataToSet)
	} else if inputDataToSet.Type().ConvertibleTo(fieldVal.Type()) && (inputDataToSet.Type().Kind() == reflect.Float64) {
		fieldVal.Set(inputDataToSet.Convert(fieldVal.Type()))
	} else {
		return fmt.Errorf("do not set field %s to %s", fieldName, inputDataToSet)
	}
	return nil
}
