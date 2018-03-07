package mw

import (
	"reflect"
	"time"
)

func copyDirectly(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	orig := reflect.ValueOf(src)
	dest := reflect.New(orig.Type()).Elem()
	copyRecursive(orig, dest)
	return dest.Interface()
}

func copyRecursive(orig, dest reflect.Value) {
	switch orig.Kind() {
	case reflect.Ptr:
		origValue := orig.Elem()

		if !origValue.IsValid() {
			return
		}
		dest.Set(reflect.New(origValue.Type()))
		copyRecursive(origValue, dest.Elem())
	case reflect.Interface:
		if orig.IsNil() {
			return
		}
		origValue := orig.Elem()

		copyValue := reflect.New(origValue.Type()).Elem()
		copyRecursive(origValue, copyValue)
		dest.Set(copyValue)
	case reflect.Struct:
		t, ok := orig.Interface().(time.Time)
		if ok {
			dest.Set(reflect.ValueOf(t))
			return
		}
		for i := 0; i < orig.NumField(); i++ {
			if orig.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(orig.Field(i), dest.Field(i))
		}
	case reflect.Slice:
		if orig.IsNil() {
			return
		}
		dest.Set(reflect.MakeSlice(orig.Type(), orig.Len(), orig.Cap()))
		for i := 0; i < orig.Len(); i++ {
			copyRecursive(orig.Index(i), dest.Index(i))
		}
	case reflect.Map:
		if orig.IsNil() {
			return
		}
		dest.Set(reflect.MakeMap(orig.Type()))
		for _, key := range orig.MapKeys() {
			origValue := orig.MapIndex(key)
			copyValue := reflect.New(origValue.Type()).Elem()
			copyRecursive(origValue, copyValue)
			copyKey := copyDirectly(key.Interface())
			dest.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
		}
	default:
		dest.Set(orig)
	}
}
