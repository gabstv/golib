package nan2z

import (
	"math"
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

func setDefaultValue(v reflect.Value, defaultValue float64, hasnan *bool) bool {
	kinder := v.Kind()
	if isFloat(kinder) {
		if val := v.Float(); math.IsNaN(val) {
			*hasnan = true
			if v.CanSet() {
				v.SetFloat(defaultValue)
				return true
			}
			return false
		}
	}
	if isInt(kinder) {
		if val := float64(v.Int()); math.IsNaN(val) {
			*hasnan = true
			if v.CanSet() {
				v.SetInt(int64(defaultValue))
				return true
			}
			return false
		}
	}
	if isUint(kinder) {
		if val := float64(v.Uint()); math.IsNaN(val) {
			*hasnan = true
			if v.CanSet() {
				v.SetUint(uint64(defaultValue))
				return true
			}
			return false
		}
	}
	return true
}

func run(target reflect.Value, defaultValue float64, hasnan, setpass *bool) {
	switch target.Kind() {
	case reflect.Array, reflect.Slice:
		if target.Len() > 0 {
			for i := 0; i < target.Len(); i++ {
				if target.Index(i).Kind() == reflect.Ptr {
					// loop through pointers and keep such recursive wow
					run(target.Index(i).Elem(), defaultValue, hasnan, setpass)
				} else if target.Index(i).Kind() == reflect.Interface {
					run(reflect.ValueOf(target.Index(i).Interface()), defaultValue, hasnan, setpass)
				} else if isNumeric(target.Index(i).Kind()) {
					// loop through and purge NaNs
					couldset := setDefaultValue(target.Index(i), defaultValue, hasnan)
					if !couldset {
						*setpass = couldset
					}
				} else {
					run(target.Index(i), defaultValue, hasnan, setpass)
				}
			}
		}
	case reflect.Map:
		keys := target.MapKeys()
		if len(keys) > 0 {
			for i := 0; i < len(keys); i++ {
				v := target.MapIndex(keys[i])
				if v.Kind() == reflect.Ptr {
					run(v.Elem(), defaultValue, hasnan, setpass)
				} else if v.Kind() == reflect.Interface {
					run(reflect.ValueOf(v.Interface()), defaultValue, hasnan, setpass)
				} else if isNumeric(v.Kind()) {
					couldset := setDefaultValue(v, defaultValue, hasnan)
					if !couldset {
						*setpass = couldset
					}
				} else {
					run(v, defaultValue, hasnan, setpass)
				}
			}
		}
	case reflect.Struct:
		flen := target.NumField()
		for i := 0; i < flen; i++ {
			v := target.Field(i)
			if v.Kind() == reflect.Ptr {
				run(v.Elem(), defaultValue, hasnan, setpass)
			} else if v.Kind() == reflect.Interface {
				run(reflect.ValueOf(v.Interface()), defaultValue, hasnan, setpass)
			} else if isNumeric(v.Kind()) {
				couldset := setDefaultValue(v, defaultValue, hasnan)
				if !couldset {
					*setpass = couldset
				}
			} else {
				run(v, defaultValue, hasnan, setpass)
			}
		}
	case reflect.Ptr:
		run(target.Elem(), defaultValue, hasnan, setpass)
	}
	if isNumeric(target.Kind()) {
		couldset := setDefaultValue(target, defaultValue, hasnan)
		if !couldset {
			*setpass = couldset
		}
	}
}

func Run(target interface{}) (hasnan, success bool) {
	success = true
	run(reflect.ValueOf(target), 0, &hasnan, &success)
	return
}
