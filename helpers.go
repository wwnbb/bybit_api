package qant_api_bybit

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func ptr[T any](value T) (pointer *T) { return &value }

func PrettyPrint(v interface{}) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println(string(b))
	default:
		fmt.Printf("%+v\n", v)
	}
}
