package notstd

import (
	"encoding/json"
	"fmt"
)

func PrettyJSON[T any](v T) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshaling %T to json: %s", v, err)
	}
	return string(data)
}
