package model

// TodoState represents the state of a todo item
type TodoState string

const (
	StateTODO  TodoState = "TODO"
	StatePROG  TodoState = "PROG"
	StateBLOCK TodoState = "BLOCK"
	StateDONE  TodoState = "DONE"
	StateNone  TodoState = ""
)
