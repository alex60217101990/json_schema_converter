package enums

// PatchOperation is an enumeration of json patch operation type values
type PatchOperation int

const (
	Replace PatchOperation = iota // PatchOperation replace json field
	Remove                        // PatchOperation remove json field
	Add                           // PatchOperation add json field
)
