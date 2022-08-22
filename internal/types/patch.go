package types

import (
	"encoding/json"

	"github.com/alex60217101990/json_schema_generator/internal/enums"
)

type Patch struct {
	OperationType enums.PatchOperation `json:"op"`
	Path          string               `json:"path"`
	Value         json.RawMessage      `json:"value"`
}

type Required struct {
	RequiredList []string `json:"required"`
}
