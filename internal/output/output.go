package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
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

	headers := rows[0].Headers
	var builder strings.Builder
	_, _ = fmt.Fprintf(&builder, "[%d]{%s}:\n", len(rows), strings.Join(headers, ","))
	for _, row := range rows {
		values := make([]string, 0, len(headers))
		for _, key := range headers {
			values = append(values, encodeValue(row.Values[key]))
		}
		builder.WriteString("  " + strings.Join(values, ",") + "\n")
	}

	_, err = io.WriteString(writer, builder.String())
	return err
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
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return rowData{Headers: []string{}, Values: map[string]any{}}, nil
		}
		value = value.Elem()
	}

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

func encodeValue(value any) string {
	switch typed := value.(type) {
	case string:
		if strings.ContainsAny(typed, ",\"\n") {
			return strconv.Quote(typed)
		}
		return typed
	case fmt.Stringer:
		return encodeValue(typed.String())
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(value)
	}
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
