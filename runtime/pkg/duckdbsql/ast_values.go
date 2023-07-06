package duckdbsql

import "strings"

// TODO: figure out a way to cast map[string]interface{} returned by json unmarshal to map[astNodeKey]interface{} and replace string in key to astNodeKey
type astNode map[string]interface{}

const (
	astKeyError            string = "error"
	astKeyErrorMessage     string = "error_message"
	astKeyStatements       string = "statements"
	astKeyNode             string = "node"
	astKeyType             string = "type"
	astKeyKey              string = "key"
	astKeyFromTable        string = "from_table"
	astKeySelectColumnList string = "select_list"
	astKeyTableName        string = "table_name"
	astKeyFunction         string = "function"
	astKeyFunctionName     string = "function_name"
	astKeyChildren         string = "children"
	astKeyValue            string = "value"
	astKeyLeft             string = "left"
	astKeyRight            string = "right"
	astKeyColumnNames      string = "column_names"
	astKeyAlias            string = "alias"
	astKeyID               string = "id"
	astKeySample           string = "sample"
	astKeyColumnNameAlias  string = "column_name_alias"
	astKeyModifiers        string = "modifiers"
	astKeyLimit            string = "limit"
	astKeyClass            string = "class"
	astKeyCTE              string = "cte_map"
	astKeyMap              string = "map"
	astKeyQuery            string = "query"
	astKeySubQuery         string = "subquery"
	astKetRelationName     string = "relation_name"
)

func toBoolean(a astNode, k string) bool {
	v, ok := a[k]
	if !ok {
		return false
	}
	switch vt := v.(type) {
	case bool:
		return vt
	default:
		return false
	}
}

func toString(a astNode, k string) string {
	v, ok := a[k]
	if !ok {
		return ""
	}
	switch vt := v.(type) {
	case string:
		return vt
	default:
		return ""
	}
}

func toNode(a astNode, k string) astNode {
	v, ok := a[k]
	if !ok {
		return nil
	}
	switch vt := v.(type) {
	case map[string]interface{}:
		return vt
	default:
		return nil
	}
}

func toArray(a astNode, k string) []interface{} {
	v, ok := a[k]
	if !ok {
		return make([]interface{}, 0)
	}
	switch v.(type) {
	case interface{}:
		return v.([]interface{})
	default:
		return make([]interface{}, 0)
	}
}

func toNodeArray(a astNode, k string) []astNode {
	arr := toArray(a, k)
	nodeArr := make([]astNode, len(arr))
	for i, e := range arr {
		nodeArr[i] = e.(map[string]interface{})
	}
	return nodeArr
}

func toTypedArray[E interface{}](a astNode, k string) []E {
	arr := toArray(a, k)
	typedArr := make([]E, len(arr))
	for i, e := range arr {
		typedArr[i] = e.(E)
	}
	return typedArr
}

func getColumnName(node astNode) string {
	alias := toString(node, astKeyAlias)
	if alias != "" {
		return alias
	}
	return strings.Join(toTypedArray[string](node, astKeyColumnNames), ".")
}
