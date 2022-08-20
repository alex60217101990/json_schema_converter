package types

import (
	"encoding/json"
	"github.com/ccs-installer/uber-installer/src/dind-pipeline-installer/schema-generator/internal/enums"
)

type Patch struct {
	OperationType enums.PatchOperation `json:"op"`
	Path          string               `json:"path"`
	Value         json.RawMessage      `json:"value"`
}
