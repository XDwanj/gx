package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	toon "github.com/toon-format/toon-go"
)

type rowData struct {
	Headers []string
	Values  map[string]any
}

func PrintJSON(writer io.Writer, value any) error {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("gx: json encoding error: %w", err)
	}
	_, err = fmt.Fprintln(writer, string(bytes))
	return err
}

func PrintTOON(writer io.Writer, value any) error {
	rows, err := normalizeRows(value)
	if err != nil {
		return fmt.Errorf("gx: toon encoding error: %w", err)
	}
	if len(rows) == 0 {
		_, err = io.WriteString(writer, "[0]{}:\n")
		return err
	}

	headers := collectHeaders(rows)
	tabularRows := make([]toon.Object, 0, len(rows))
	for _, row := range rows {
		tabularRows = append(tabularRows, toTOONRow(headers, row.Values))
	}

	encoded, err := toon.Marshal(tabularRows)
	if err != nil {
		return fmt.Errorf("gx: toon encoding error: %w", err)
	}
	_, err = fmt.Fprintln(writer, string(encoded))
	return err
}

func toTOONRow(headers []string, values map[string]any) toon.Object {
	fields := make([]toon.Field, 0, len(headers))
	for _, key := range headers {
		value, ok := values[key]
		if !ok {
			value = ""
		}
		fields = append(fields, toon.Field{
			Key:   key,
			Value: value,
		})
	}
	return toon.NewObject(fields...)
}

func collectHeaders(rows []rowData) []string {
	headers := make([]string, 0)
	seen := make(map[string]struct{})
	for _, row := range rows {
		for _, header := range row.Headers {
			if _, ok := seen[header]; ok {
				continue
			}
			seen[header] = struct{}{}
			headers = append(headers, header)
		}
	}
	return headers
}

func normalizeRows(value any) ([]rowData, error) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, nil
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		rows := make([]rowData, 0, rv.Len())
		for index := 0; index < rv.Len(); index++ {
			row, err := structToRow(rv.Index(index))
			if err != nil {
				return nil, err
			}
			rows = append(rows, row)
		}
		return rows, nil
	}

	row, err := structToRow(rv)
	if err != nil {
		return nil, err
	}
	return []rowData{row}, nil
}

func structToRow(value reflect.Value) (rowData, error) {
	for {
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if value.IsNil() {
				return rowData{Headers: []string{}, Values: map[string]any{}}, nil
			}
			value = value.Elem()
		default:
			goto resolved
		}
	}

resolved:
	if value.Kind() != reflect.Struct {
		return rowData{}, fmt.Errorf("unsupported toon value kind: %s", value.Kind())
	}

	headers := make([]string, 0, value.NumField())
	values := map[string]any{}
	valueType := value.Type()
	for index := 0; index < value.NumField(); index++ {
		field := valueType.Field(index)
		if !field.IsExported() {
			continue
		}

		key := field.Tag.Get("json")
		if key == "" {
			key = strings.ToLower(field.Name)
		} else {
			key = strings.Split(key, ",")[0]
		}
		if key == "-" || key == "" {
			continue
		}

		fieldValue := value.Field(index)
		if strings.Contains(field.Tag.Get("json"), "omitempty") && isEmptyValue(fieldValue) {
			continue
		}

		headers = append(headers, key)
		values[key] = fieldValue.Interface()
	}

	return rowData{Headers: headers, Values: values}, nil
}

func isEmptyValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return value.IsNil()
	default:
		return false
	}
}
