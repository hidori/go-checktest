package example

type Example struct {
	Value string
}

func NewExample() *Example {
	return &Example{}
}

func NewExampleWithValue(value string) *Example {
	return &Example{
		Value: value,
	}
}
