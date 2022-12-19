package queries

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

func quoteName(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

func escapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func escapeDoubleQuotes(column string) string {
	return strings.ReplaceAll(column, "\"", "\"\"")
}

func safeName(name string) string {
	return quoteName(escapeDoubleQuotes(name))
}

func dropTempTable(olap drivers.OLAPStore, priority int, tableName string) {
	rs, er := olap.Execute(context.Background(), &drivers.Statement{
		Query:    `DROP TABLE "` + tableName + `"`,
		Priority: priority,
	})
	if er == nil {
		rs.Close()
	}
}

func replaceHyphen(column string) string {
	return strings.ReplaceAll(column, "-", "_")
}
