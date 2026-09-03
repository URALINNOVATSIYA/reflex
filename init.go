package reflex

import "reflect"

func init() {
	field, _ := reflect.TypeFor[reflect.Value]().FieldByName("flag")
	flagOffset = field.Offset
}
