package json

import (
	"bytes"
	"encoding/json"
	"strconv"
)

const hex = "0123456789abcdef"

var noEscape = [256]bool{}

func init() {
	// the idea is to fill up with characters from 0 to ~, and indicate whether or not it is valid.
	// So, when processing if we receive a non-valid JSON character, we can escape it with the correct character
	for i := 0; i <= 0x7e; i++ {
		noEscape[i] = i >= 0x20 && i != '\\' && i != '"'
	}
}

// Encoder JSON encoder
type Encoder struct{}

// BeginObj begins the JSON object
func (Encoder) BeginObj(dst []byte) []byte {
	return append(dst, '{')
}

// EndObj ends the JSON object
func (Encoder) EndObj(dst []byte) []byte {
	return append(dst, '}')
}

// NewLine adds a newline
func (Encoder) NewLine(dst []byte) []byte {
	return append(dst, '\n')
}

// Comma adds a trailing comma
func (Encoder) Comma(dst []byte) []byte {
	return append(dst, ',')
}

// Append adds the given value
func (Encoder) Append(dst, value []byte) []byte {
	return append(dst, value...)
}

// BeginArray begins an array
func (Encoder) BeginArray(dst []byte) []byte {
	return append(dst, '[')
}

// EndArray ends the array
func (Encoder) EndArray(dst []byte) []byte {
	return append(dst, ']')
}

// ObjKey adds the given key
func (Encoder) ObjKey(dst []byte, key string) []byte {
	dst = append(dst, '"')
	dst = append(dst, key...)
	dst = append(dst, '"')
	return append(dst, ':')
}

// ValueBytes adds the given value
func (Encoder) ValueBytes(dst, value []byte) []byte {
	dst = append(dst, '"')
	for _, b := range value {
		if !noEscape[b] {
			dst = escapeValue(dst, b)
			continue
		}

		dst = append(dst, b)
	}

	return append(dst, '"')
}

// ValueString adds the given value
func (Encoder) ValueString(dst []byte, value string) []byte {
	dst = append(dst, '"')
	for _, b := range []byte(value) {
		if !noEscape[b] {
			dst = escapeValue(dst, b)
			continue
		}

		dst = append(dst, b)
	}

	return append(dst, '"')
}

// ValueMarshaler addds a Marshaler
func (e Encoder) ValueMarshaler(dst []byte, value json.Marshaler) ([]byte, error) {
	val, err := value.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	if err = json.Compact(&result, val); err != nil {
		return nil, err
	}

	return e.Append(dst, result.Bytes()), nil
}

// ValueInt adds the given value int
func (Encoder) ValueInt(dst []byte, value int) []byte {
	return strconv.AppendInt(dst, int64(value), 10)
}

// ValueUint adds the given value int
func (Encoder) ValueUint(dst []byte, value uint) []byte {
	return strconv.AppendUint(dst, uint64(value), 10)
}

// ValueInterface adds the given value
// nolint
func (e Encoder) ValueInterface(dst []byte, valueIf interface{}) (result []byte, err error) {
	switch value := valueIf.(type) {
	case string:
		dst = e.ValueString(dst, value)
	case []byte:
		dst = e.ValueBytes(dst, value)
	case int:
		dst = e.ValueInt(dst, value)
	case int8:
		dst = e.ValueInt(dst, int(value))
	case int16:
		dst = e.ValueInt(dst, int(value))
	case int32:
		dst = e.ValueInt(dst, int(value))
	case int64:
		dst = e.ValueInt(dst, int(value))
	case uint:
		dst = e.ValueUint(dst, uint(value))
	case uint8:
		dst = e.ValueUint(dst, uint(value))
	case uint16:
		dst = e.ValueUint(dst, uint(value))
	case uint32:
		dst = e.ValueUint(dst, uint(value))
	case uint64:
		dst = e.ValueUint(dst, uint(value))
	case json.Marshaler:
		dst, err = e.ValueMarshaler(dst, value)
	default:
		val, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		dst = e.Append(dst, val)
	}

	return dst, err
}

func escapeValue(dst []byte, b byte) []byte {
	switch b {
	case '"', '\\':
		dst = append(dst, '\\', b)
	case '\b':
		dst = append(dst, '\\', 'b')
	case '\f':
		dst = append(dst, '\\', 'f')
	case '\n':
		dst = append(dst, '\\', 'n')
	case '\r':
		dst = append(dst, '\\', 'r')
	case '\t':
		dst = append(dst, '\\', 't')
	default:
		dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
	}

	return dst
}
