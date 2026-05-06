package sdk

// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

// ABIVersion is the current WASM ABI version.
// Extensions must export _abi_version() returning this value if they want
// strict version checking (optional in v1).
const ABIVersion = 1

// Error codes returned by host_call.
const (
	ErrOK      int32 = 0
	ErrGeneral int32 = 1
	ErrCancel  int32 = 2
)
