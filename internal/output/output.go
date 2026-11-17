package output

import (
	"encoding/json"
	"fmt"
	"io"

	toon "github.com/toon-format/toon-go"
)

func PrintJSON(writer io.Writer, value any) error {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("gx: json encoding error: %w", err)
	}
	_, err = fmt.Fprintln(writer, string(bytes))
	return err
}

func PrintTOON(writer io.Writer, value any) error {
	encoded, err := toon.Marshal(value)
	if err != nil {
		return fmt.Errorf("gx: toon encoding error: %w", err)
	}
	_, err = fmt.Fprintln(writer, string(encoded))
	return err
}
