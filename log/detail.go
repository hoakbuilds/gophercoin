package log

const (
	field = iota
)

// Detail specifies the detail that should be shown during logging
type Detail struct {
	key        string
	value      interface{}
	detailType int
}

// NewDetail returns a new Detail
func NewDetail(key string, value interface{}) Detail {
	return NewField(key, value)
}

// NewField returns a new field
func NewField(key string, value interface{}) Detail {
	return Detail{key, value, field}
}
