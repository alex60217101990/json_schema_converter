package utils

import (
	"bytes"
	"encoding/json"
)

func PrettyString(str []byte) ([]byte, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, str, "", "\t"); err != nil {
		return nil, err
	}
	return prettyJSON.Bytes(), nil
}
