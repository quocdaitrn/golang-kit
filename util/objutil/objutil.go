package objutil

import (
	"reflect"
)

// Map maps object's data from source type to destination type by field name.
func Map(src interface{}, dst interface{}) {
	fromVal := reflect.ValueOf(src)
	toVal := reflect.ValueOf(dst)

	if (fromVal.Kind() != reflect.Ptr || fromVal.Elem().Kind() != reflect.Struct) && fromVal.Kind() != reflect.Struct {
		panic("src argument must be a struct address or a struct")
	}
	if toVal.Kind() != reflect.Ptr || toVal.Elem().Kind() != reflect.Struct {
		panic("dst argument must be a struct address")
	}

	if fromVal.Kind() == reflect.Ptr {
		fromVal = reflect.Indirect(fromVal)
	}
	toVal = reflect.Indirect(toVal)

	for i := 0; i < fromVal.NumField(); i++ {
		fFieldTyp := fromVal.Type().Field(i)
		fFieldVal := fromVal.FieldByName(fFieldTyp.Name)
		tFieldVal := toVal.FieldByName(fFieldTyp.Name)

		if !tFieldVal.IsValid() || !tFieldVal.CanSet() {
			continue
		}

		if fFieldVal.Kind() == reflect.Interface {
			fFieldVal = fFieldVal.Elem()
		}

		if tFieldVal.Kind() == reflect.Interface {
			tFieldVal.Set(fFieldVal)
		} else if (fFieldVal.Kind() == reflect.Ptr && tFieldVal.Kind() == reflect.Ptr) ||
			(fFieldVal.Kind() != reflect.Ptr && tFieldVal.Kind() != reflect.Ptr && fFieldVal.Kind() != reflect.Interface) {
			if fFieldVal.Type() == tFieldVal.Type() {
				tFieldVal.Set(fFieldVal)
			}
		} else {
			if fFieldVal.Kind() == reflect.Ptr {
				if !fFieldVal.IsNil() {
					fFieldVal = reflect.Indirect(fFieldVal)
					if fFieldVal.Type() == tFieldVal.Type() {
						tFieldVal.Set(fFieldVal)
					}
				}
			} else if tFieldVal.Kind() == reflect.Ptr {
				if tFieldVal.Type().Elem() == fFieldVal.Type() {
					pNewTo := reflect.New(tFieldVal.Type().Elem())
					newTo := reflect.Indirect(pNewTo)
					newTo.Set(fFieldVal)
					tFieldVal.Set(pNewTo)
				}
			}
		}
	}
}

// SlicePtrMap map data of objects slice from source type to destination type
// by field name.
func SlicePtrMap(src interface{}, dst interface{}) {
	typFrom := reflect.TypeOf(src)
	typTo := reflect.TypeOf(dst)

	if !(typFrom.Kind() == reflect.Ptr && typFrom.Elem().Kind() == reflect.Slice) {
		panic("src argument must be a slice")
	}
	if !(typTo.Kind() == reflect.Ptr && typTo.Elem().Kind() == reflect.Slice) {
		panic("src argument must be a slice")
	}

	sliceFrom := reflect.ValueOf(src).Elem()
	sliceTo := reflect.MakeSlice(typTo.Elem(), 0, sliceFrom.Len())

	for i := 0; i < sliceFrom.Len(); i++ {
		elemFrom := sliceFrom.Index(i).Addr().Interface()
		elemTo := reflect.New(typTo.Elem()).Interface()
		Map(elemFrom, elemTo)
		sliceTo = reflect.Append(sliceTo, reflect.ValueOf(elemTo).Elem())
	}
	reflect.ValueOf(dst).Elem().Set(sliceTo)
}

// FlattenFieldNames returns a map with keys that are filed could be access directly
// by field name.
func FlattenFieldNames(typ reflect.Type) map[string]bool {

	names := map[string]bool{}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		fieldTyp := typ.Field(i)
		if fieldTyp.Anonymous {
			anonymousName := FlattenFieldNames(fieldTyp.Type)
			for n, v := range anonymousName {
				names[n] = v
			}
		} else {
			names[fieldTyp.Name] = true
		}
	}
	return names
}

// IsZero check if value of input object is zero value.
func IsZero(i interface{}) bool {
	zero := reflect.Zero(reflect.TypeOf(i)).Interface()
	return reflect.DeepEqual(i, zero)
}
