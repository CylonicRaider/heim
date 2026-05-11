package psql

import (
	"fmt"
	"reflect"
	"strings"

	"euphoria.leet.nu/lib/scope"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
	"gopkg.in/gorp.v1"

	"euphoria.leet.nu/heim/proto/logging"
)

func rollback(ctx scope.Context, t *gorp.Transaction) {
	if err := t.Rollback(); err != nil {
		logging.Logger(ctx).Printf("rollback error: %s", err)
	}
}

func allColumns(dbMap *gorp.DbMap, row interface{}, prefix string, aliases ...string) (string, error) {
	aliasMap := map[string]string{}
	for i := 0; i+1 < len(aliases); i += 2 {
		aliasMap[aliases[i]] = aliases[i+1]
	}

	tableMap, err := dbMap.TableFor(reflect.TypeOf(row), false)
	if err != nil {
		return "", err
	}

	parts := make([]string, 0, len(tableMap.Columns))
	for _, col := range tableMap.Columns {
		if !col.Transient {
			part := dbMap.Dialect.QuoteField(col.ColumnName)
			if prefix != "" {
				part = fmt.Sprintf("%s.%s", prefix, part)
			}
			if alias, ok := aliasMap[col.ColumnName]; ok {
				part = fmt.Sprintf("%s AS %s", part, dbMap.Dialect.QuoteField(alias))
			}
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ", "), nil
}

func isUniqueViolation(err error) bool {
	return pq.As(err, pqerror.UniqueViolation) != nil
}
