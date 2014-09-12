package mq

// based on https://github.com/jmoiron/jsonq/blob/master/jsonq.go

import (
	"errors"
	"log"
	"reflect"
)

var (
	ERR_NOT_FOUND        = errors.New("not found")
	ERR_INDEX_NOT_STRING = errors.New("first query item must be a string")
	ERR_INDEX_NOT_VALID  = errors.New("query item must be string or int")
	ERR_CONVERT          = errors.New("cannot convert")
)

type QMap map[string]interface{}

func (q QMap) C(index ...interface{}) *QMapChild {
	v, err := q.CV(index...)
	if err != nil {
		return &QMapChild{nil, err}
	}
	return v
}

func (q QMap) CV(index ...interface{}) (*QMapChild, error) {
	val0 := reflect.ValueOf(q)
	log.Println(val0.Kind())
	//
	if len(index) < 1 {
		return nil, ERR_NOT_FOUND
	}
	if reflect.ValueOf(index[0]).Kind() != reflect.String {
		return nil, ERR_INDEX_NOT_STRING
	}
	v0, ok0 := q[index[0].(string)]
	if !ok0 {
		return nil, ERR_NOT_FOUND
	}
	if len(index) == 1 {
		return newChild(v0), nil
	}
	//
	var val interface{}
	//
	val = v0
	for _, v := range index[1:] {
		vv := reflect.ValueOf(v)
		for vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
		}
		switch vv.Kind() {
		case reflect.String:
			//
			mapv := reflect.ValueOf(val)
			for mapv.Kind() == reflect.Ptr {
				mapv = mapv.Elem()
			}
			switch mapv.Kind() {
			case reflect.Map, reflect.Struct:
				//as expected
				val = mapv.MapIndex(vv)
			default:
				return nil, errors.New("can't search for a string key in a non struct non map value " + mapv.Kind().String())
			}
			//
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			//
			slicev := reflect.ValueOf(val)
			for slicev.Kind() == reflect.Ptr {
				slicev = slicev.Elem()
			}
			switch slicev.Kind() {
			case reflect.Slice:
				//as expected
				val = slicev.Index(int(vv.Int()))
			default:
				return nil, errors.New("can't search for an int index in a non slice value " + slicev.Kind().String())
			}
			//
		default:
			return nil, ERR_INDEX_NOT_VALID
		}
	}
	return newChild(val), nil
}

func newChild(val interface{}) *QMapChild {
	return &QMapChild{val, nil}
}

type QMapChild struct {
	obj interface{}
	err error
}

func (c *QMapChild) BoolV() (bool, error) {
	if c.err != nil {
		return false, c.err
	}
	vv := reflect.ValueOf(c.obj)
	if vv.Kind() != reflect.Bool {
		return false, ERR_CONVERT
	}
	return vv.Bool(), nil
}

func (c *QMapChild) Bool() bool {
	v, _ := c.BoolV()
	return v
}

func (c *QMapChild) IntV() (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	vv := getval(reflect.ValueOf(c.obj))
	switch vv.Kind() {
	case reflect.Float64, reflect.Float32:
		return int(vv.Float()), nil
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		return int(vv.Int()), nil
	}
	return 0, ERR_CONVERT
}

func (c *QMapChild) Int() int {
	v, _ := c.IntV()
	return v
}

func (c *QMapChild) F64V() (float64, error) {
	if c.err != nil {
		return 0, c.err
	}
	vv := getval(reflect.ValueOf(c.obj))
	switch vv.Kind() {
	case reflect.Float64, reflect.Float32:
		return vv.Float(), nil
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		return float64(vv.Int()), nil
	}
	return 0, ERR_CONVERT
}

func (c *QMapChild) F64() float64 {
	v, _ := c.F64V()
	return v
}

func (c *QMapChild) StrV() (string, error) {
	if c.err != nil {
		return "", c.err
	}
	vv := getval(reflect.ValueOf(c.obj))
	switch vv.Kind() {
	case reflect.String:
		return vv.String()
	}
	return "", ERR_CONVERT
}

func (c *QMapChild) Str() string {
	v, _ := c.StrV()
	return v
}

func getval(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
