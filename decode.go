package ini

import (
	"reflect"
	"strconv"
)

// An UnmarshalTypeError describes a value that was not appropriate for a value
// of a specific Go type.
type UnmarshalTypeError struct {
	Value  string       // description of value - "bool", "array", "number -5"
	Type   reflect.Type // type of Go value it could not be assigned to
	Struct string       // name of the struct type containing the field
	Field  string       // name of the field within the struct
}

func (e *UnmarshalTypeError) Error() string {
	if e.Struct != "" || e.Field != "" {
		return "ini: cannot unmarshal " + e.Value + " into Go struct field " + e.Struct + "." + e.Field + " of type " + e.Type.String()
	}
	return "ini: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
}

func Unmarshal(data []byte, v interface{}) error {
	return nil
}

// decode sets the underlying values of the value to which rv points to the
// concrete value stored in the corresponding field of ast.
func decode(ast map[string]section, rv reflect.Value) error {
	if rv.Type().Kind() != reflect.Ptr {
		return &UnmarshalTypeError{
			Value: reflect.ValueOf(ast).String(),
			Type:  rv.Type(),
		}
	}

	rv = reflect.Indirect(rv)
	if rv.Type().Kind() != reflect.Struct {
		return &UnmarshalTypeError{
			Value: reflect.ValueOf(ast).String(),
			Type:  rv.Type(),
		}
	}

	for i := 0; i < rv.NumField(); i++ {
		sf := rv.Type().Field(i)

		t := newTag(sf)
		if t.name == "-" {
			continue
		}

		switch sf.Type.Kind() {
		case reflect.Struct:
			sv := rv.Field(i).Addr()
			val := ast[t.name]
			if err := decodeStruct(val, sv); err != nil {
				return err
			}
		case reflect.String:
			sv := rv.Field(i).Addr()
			val := ast[""].props[t.name].val[0]
			if err := decodeString(val, sv); err != nil {
				return err
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sv := rv.Field(i).Addr()
			val := ast[""].props[t.name].val[0]
			if err := decodeInt(val, sv); err != nil {
				return err
			}
		}
	}

	return nil
}

// decodeStruct sets the underlying values of the fields of the value to which
// rv points to the concrete values stored in i. If rv is not a reflect.Ptr,
// decodeStruct returns UnmarshalTypeError.
func decodeStruct(i interface{}, rv reflect.Value) error {
	if reflect.TypeOf(i) != reflect.TypeOf(section{}) || rv.Type().Kind() != reflect.Ptr {
		return &UnmarshalTypeError{
			Value: reflect.ValueOf(i).String(),
			Type:  rv.Type(),
		}
	}

	s := i.(section)
	rv = rv.Elem()

	/* magic */
	for i := 0; i < rv.NumField(); i++ {
		sf := rv.Type().Field(i)

		t := newTag(sf)
		if t.name == "-" {
			continue
		}

		switch sf.Type.Kind() {
		case reflect.Slice:
			sv := rv.Field(i).Addr()
			val := s.props[t.name]
			if err := decodeSlice(val, sv); err != nil {
				return err
			}
		case reflect.String:
			sv := rv.Field(i).Addr()
			val := s.props[t.name].val[0]
			if err := decodeString(val, sv); err != nil {
				return err
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sv := rv.Field(i).Addr()
			val := s.props[t.name].val[0]
			if err := decodeInt(val, sv); err != nil {
				return err
			}
		}
	}

	return nil
}

// decodeSlice sets the underlying values of the elements of the value to which
// rv points to the concrete values stored in i.
func decodeSlice(i interface{}, rv reflect.Value) error {
	if reflect.TypeOf(i) != reflect.TypeOf(property{}) || rv.Type().Kind() != reflect.Ptr {
		return &UnmarshalTypeError{
			Value: reflect.ValueOf(i).String(),
			Type:  rv.Type(),
		}
	}

	p := i.(property)
	rv = rv.Elem()

	var decoderFunc func(interface{}, reflect.Value) error

	switch rv.Type().Elem().Kind() {
	case reflect.String:
		decoderFunc = decodeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		decoderFunc = decodeInt
	default:
		return &UnmarshalTypeError{
			Value: reflect.ValueOf(i).String(),
			Type:  rv.Type(),
		}
	}

	vv := reflect.MakeSlice(rv.Type(), len(p.val), cap(p.val))

	for i := 0; i < vv.Len(); i++ {
		sv := vv.Index(i).Addr()
		val := p.val[i]
		if err := decoderFunc(val, sv); err != nil {
			return err
		}
	}

	rv.Set(vv)

	return nil
}

// decodeString sets the underlying value of the value to which rv points to
// the concrete value stored in i. If rv is not a reflect.Ptr, decodeString
// returns UnmarshalTypeError.
func decodeString(i interface{}, rv reflect.Value) error {
	if reflect.TypeOf(i).Kind() != reflect.String || rv.Type().Kind() != reflect.Ptr {
		return &UnmarshalTypeError{
			Value: reflect.ValueOf(i).String(),
			Type:  rv.Type(),
		}
	}

	rv.Elem().SetString(i.(string))
	return nil
}

// decodeInt sets the underlying value of the value to which rv points to the
// concrete value stored in i. If rv is not a reflect.Ptr, decodeInt returns
// UnmarshalTypeError.
func decodeInt(i interface{}, rv reflect.Value) error {
	if reflect.TypeOf(i).Kind() != reflect.String || rv.Type().Kind() != reflect.Ptr {
		return &UnmarshalTypeError{
			Value: reflect.ValueOf(i).String(),
			Type:  rv.Type(),
		}
	}

	n, err := strconv.ParseInt(i.(string), 10, 64)
	if err != nil {
		switch err.(*strconv.NumError).Err {
		case strconv.ErrRange:
		default:
			return &UnmarshalTypeError{
				Value: reflect.ValueOf(i).String(),
				Type:  rv.Type(),
			}
		}
	}

	rv.Elem().SetInt(n)
	return nil
}
