package model

type KeyValuePair struct {
	Key   string
	Value string
}

type SelectOption struct {
	Name  string
	Value string
}

type ComponentTrigger struct {
	ID           string
	Class        string
	ComponentUrl string
	Trigger      string
	WithCodeView bool
}
