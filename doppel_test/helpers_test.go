package doppel_test

import (
	"reflect"
	"testing"
)

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func reflectDeepClone[T any](src T) T {
	cloned := cloneReflectValue(reflect.ValueOf(src))

	return cloned.Interface().(T)
}

func cloneReflectValue(src reflect.Value) reflect.Value {
	if !src.IsValid() {
		return src
	}

	switch src.Kind() {
	case reflect.Pointer:
		return cloneReflectPointer(src)

	case reflect.Interface:
		return cloneReflectInterface(src)

	case reflect.Struct:
		return cloneReflectStruct(src)

	case reflect.Slice:
		return cloneReflectSlice(src)

	case reflect.Array:
		return cloneReflectArray(src)

	case reflect.Map:
		return cloneReflectMap(src)

	default:
		return src
	}
}

func cloneReflectPointer(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return reflect.Zero(src.Type())
	}

	clonedValue := cloneReflectValue(src.Elem())
	clonedPointer := reflect.New(src.Type().Elem())
	clonedPointer.Elem().Set(clonedValue)

	return clonedPointer
}

func cloneReflectInterface(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return reflect.Zero(src.Type())
	}

	clonedValue := cloneReflectValue(src.Elem())

	return clonedValue.Convert(src.Elem().Type())
}

func cloneReflectStruct(src reflect.Value) reflect.Value {
	cloned := reflect.New(src.Type()).Elem()

	for i := range src.NumField() {
		sourceField := src.Field(i)
		targetField := cloned.Field(i)

		if !targetField.CanSet() {
			continue
		}

		targetField.Set(cloneReflectValue(sourceField))
	}

	return cloned
}

func cloneReflectSlice(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return reflect.Zero(src.Type())
	}

	cloned := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())

	for i := range src.Len() {
		cloned.Index(i).Set(cloneReflectValue(src.Index(i)))
	}

	return cloned
}

func cloneReflectArray(src reflect.Value) reflect.Value {
	cloned := reflect.New(src.Type()).Elem()

	for i := range src.Len() {
		cloned.Index(i).Set(cloneReflectValue(src.Index(i)))
	}

	return cloned
}

func cloneReflectMap(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return reflect.Zero(src.Type())
	}

	cloned := reflect.MakeMapWithSize(src.Type(), src.Len())

	for _, key := range src.MapKeys() {
		clonedKey := cloneReflectValue(key)
		clonedValue := cloneReflectValue(src.MapIndex(key))

		cloned.SetMapIndex(clonedKey, clonedValue)
	}

	return cloned
}
