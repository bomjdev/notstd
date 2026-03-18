package notstd

import (
	"encoding/json"
	"fmt"
)

func PrettyJSON[T any](v T) string {
	return PrettyJSONIndent(v, "\t")
}

func PrettyJSONIndent[T any](v T, indent string) string {
	data, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return fmt.Sprintf("error marshaling %T to json: %s", v, err)
	}
	return string(data)
}

func PrettifyRawJSON(data []byte) (string, error) {
	var v json.RawMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	return PrettyJSON(v), nil
}
