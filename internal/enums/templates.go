package enums

//go:generate go-enum -type=PatchOperation -transform=lower
// PatchOperation is an enumeration of json patch operation type values
type PatchOperation int

const (
	Replace PatchOperation = iota // PatchOperation replace json field
	Remove                        // PatchOperation remove json field
)
