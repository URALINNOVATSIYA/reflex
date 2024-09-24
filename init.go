package reflex

import "reflect"

func init() {
	field, _ := reflect.TypeOf(reflect.Value{}).FieldByName("flag")
	flagOffset = field.Offset
}
