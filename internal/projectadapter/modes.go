package projectadapter

// Operator mode names are explicit; absence must never select native-project.
const (
	Strict              = "strict"
	NativeProject       = "native-project"
	NativeProjectNotice = "project sources not isolated"
)
