package nan2z

import (
	"reflect"
)

func isNumeric(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}

func isInt(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

func isUint(kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isFloat(kind reflect.Kind) bool {
	switch kind {
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func isComplex(kind reflect.Kind) bool {
	switch kind {
	case reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}

func setDefaultValue(v reflect.Value, defaultValue float64) {
	kinder := v.Kind()
	if isFloat(kinder) {
		v.SetFloat(defaultValue)
	} else if isComplex(kinder) {
		v.SetComplex(x)
	}
}

func run(target interface{}, defaultValue float64) {
	v0 := reflect.ValueOf(target)
	switch v0.Kind() {
	case reflect.Array, reflect.Slice:
		if v0.Elem().Kind() == reflect.Ptr {
			// loop through pointers and keep such recursive wow

		}
		if isNumeric(v0.Elem().Kind()) {
			// loop through and purge NaNs
			for i := 0; i < v0.Len(); i++ {
				v0.Index(i)
			}
		}
	}
}
