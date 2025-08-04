package infrastructure

// import (
// 	"encoding/json"
// 	"fmt"

// 	"github.com/cjlapao/locally-cli/internal/common"
// 	"github.com/cjlapao/locally-cli/internal/environment"
// )

// func MarshalVariables(values map[string]interface{}, indentation string) string {
// 	content := ""
// 	for key, val := range values {
// 		content += fmt.Sprintf("%s = %s\n", key, marshalValue(val, 1, indentation))
// 	}

// 	return content
// }

// func marshalValue(value interface{}, level int, indent string) string {
// 	env := environment.GetInstance()
// 	switch t := value.(type) {
// 	case string:
// 		t = env.Replace(t)
// 		if common.IsDebug() {
// 			notify.Debug("Found variable of type %s with value %s", fmt.Sprintf("%T", t), fmt.Sprintf("%v", t))
// 		}
// 		return escapeString(t)

// 	case int:
// 		if common.IsDebug() {
// 			notify.Debug("Found variable of type %s with value %s", fmt.Sprintf("%T", t), fmt.Sprintf("%v", value))
// 		}
// 		return fmt.Sprintf("%v", value)
// 	case bool:
// 		if common.IsDebug() {
// 			notify.Debug("Found variable of type %s with value %s", fmt.Sprintf("%T", t), fmt.Sprintf("%v", value))
// 		}
// 		return fmt.Sprintf("%v", value)

// 	case map[string]interface{}:
// 		if common.IsDebug() {
// 			notify.Debug("Found variable of type %s with value %s", fmt.Sprintf("%T", t), fmt.Sprintf("%v", value))
// 		}

// 		if len(value.(map[string]interface{})) == 0 {
// 			return "{}"
// 		}

// 		result := "{\n"
// 		for key, val := range value.(map[string]interface{}) {
// 			parsed := marshalValue(val, level+1, indent)
// 			result += identMarshal(indent, level)
// 			result += fmt.Sprintf("%s= %v\n", key, parsed)
// 		}
// 		result += identMarshal(indent, level-1)
// 		result += "}"
// 		return result
// 	case []interface{}:
// 		if common.IsDebug() {
// 			notify.Debug("Found variable of type %s with value %s", fmt.Sprintf("%T", t), fmt.Sprintf("%v", value))
// 		}

// 		if len(value.([]interface{})) == 0 {
// 			return "[]"
// 		}

// 		result := "[\n"
// 		for i, val := range value.([]interface{}) {
// 			parsed := marshalValue(val, level+1, indent)
// 			result += identMarshal(indent, level)
// 			if i == len(value.([]interface{}))-1 {
// 				result += fmt.Sprintf("%v\n", parsed)
// 			} else {
// 				result += fmt.Sprintf("%v,\n", parsed)
// 			}
// 		}
// 		result += identMarshal(indent, level-1)
// 		result += "]"
// 		return result
// 	default:
// 		return "not supported"
// 	}
// }

// func identMarshal(indentation string, level int) string {
// 	result := ""
// 	if level <= 0 {
// 		return result
// 	}

// 	for i := 0; i < level; i++ {
// 		result += indentation
// 	}

// 	return result
// }

// func escapeString(value string) string {
// 	v, err := json.Marshal(value)
// 	if err != nil {
// 		return value
// 	}

// 	return string(v)
// }
