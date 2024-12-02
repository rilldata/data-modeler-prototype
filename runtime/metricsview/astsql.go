package metricsview

import (
	"strconv"
	"strings"
)

// SQL builds a SQL query from the AST.
// It returns the query and query arguments to be passed to the database driver.
func (a *AST) SQL() (string, []any, error) {
	b := &sqlBuilder{
		ast: a,
		out: &strings.Builder{},
	}

	if len(a.CTEs) > 0 {
		b.out.WriteString("WITH ")
		for i, cte := range a.CTEs {
			if i > 0 {
				b.out.WriteString(", ")
			}
			b.out.WriteString(cte.Alias)
			b.out.WriteString(" AS (")
			err := b.writeSelect(cte)
			if err != nil {
				return "", nil, err
			}
			b.out.WriteString(") ")
		}
	}

	var err error
	if a.query.UseDisplayNames {
		err = b.writeSelectWithDisplayNames(a.Root)
	} else {
		err = b.writeSelect(a.Root)
	}
	if err != nil {
		return "", nil, err
	}

	return b.out.String(), b.args, nil
}

type sqlBuilder struct {
	ast  *AST
	out  *strings.Builder
	args []any
}

func (b *sqlBuilder) writeSelectWithDisplayNames(n *SelectNode) error {
	b.out.WriteString("SELECT ")

	for i, f := range n.DimFields {
		displayName := f.DisplayName
		if displayName == "" {
			displayName = f.Name
		}

		if i > 0 {
			b.out.WriteString(", ")
		}
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
		b.out.WriteString(" AS ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(displayName))
	}

	for i, f := range n.MeasureFields {
		displayName := f.DisplayName
		if displayName == "" {
			displayName = f.Name
		}

		if i > 0 || len(n.DimFields) > 0 {
			b.out.WriteString(", ")
		}
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
		b.out.WriteString(" AS ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(displayName))
	}

	b.out.WriteString(" FROM (")
	err := b.writeSelect(n)
	if err != nil {
		return err
	}
	b.out.WriteString(")")

	return nil
}

func (b *sqlBuilder) writeSelect(n *SelectNode) error {
	if len(n.UnionAllSelects) > 0 {
		for i, s := range n.UnionAllSelects {
			if i > 0 {
				b.out.WriteString(" UNION ALL ")
			}
			b.out.WriteByte('(')
			err := b.writeSelect(s)
			if err != nil {
				return err
			}
			b.out.WriteByte(')')
		}
		return nil
	}

	b.out.WriteString("SELECT ")

	for i, f := range n.DimFields {
		if i > 0 {
			b.out.WriteString(", ")
		}

		b.out.WriteByte('(')
		b.out.WriteString(f.Expr)
		b.out.WriteString(") AS ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
	}

	for i, f := range n.MeasureFields {
		if i > 0 || len(n.DimFields) > 0 {
			b.out.WriteString(", ")
		}

		if f.TreatNullAs != "" {
			b.out.WriteString("COALESCE(")
		}

		b.out.WriteByte('(')
		b.out.WriteString(f.Expr)
		if f.TreatNullAs != "" {
			b.out.WriteString("), ")
			b.out.WriteString(f.TreatNullAs)
		}
		b.out.WriteString(") AS ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
	}

	if n.FromTable == nil && n.FromSelect == nil && n.FromCrossSelect == nil {
		// will happen for union all select
		return nil
	}

	b.out.WriteString(" FROM ")
	if n.FromTable != nil {
		b.out.WriteString(*n.FromTable)

		// Add unnest joins. We only and always apply these against FromTable (ensuring they are already unnested when referenced in outer SELECTs).
		for _, u := range n.Unnests {
			b.out.WriteString(", ")
			b.out.WriteString(u)
		}
	} else if n.FromSelect != nil {
		if !n.FromSelect.IsCTE {
			b.out.WriteByte('(')
			err := b.writeSelect(n.FromSelect)
			if err != nil {
				return err
			}
			b.out.WriteString(") ")
		}
		b.out.WriteString(n.FromSelect.Alias)

		for _, ljs := range n.LeftJoinSelects {
			err := b.writeJoin("LEFT", n.FromSelect, ljs)
			if err != nil {
				return err
			}
		}

		if n.SpineSelect != nil {
			err := b.writeJoin("RIGHT", n.FromSelect, n.SpineSelect)
			if err != nil {
				return err
			}
		}
		if n.JoinComparisonSelect != nil {
			err := b.writeJoin(n.JoinComparisonType, n.FromSelect, n.JoinComparisonSelect)
			if err != nil {
				return err
			}
		}
	} else if n.FromCrossSelect != nil {
		for i, s := range n.FromCrossSelect.CrossJoinSelects {
			if i > 0 {
				b.out.WriteString(" CROSS JOIN ")
			}
			b.out.WriteByte('(')
			err := b.writeSelect(s)
			if err != nil {
				return err
			}
			b.out.WriteString(") ")
			b.out.WriteString(s.Alias)
		}
	} else {
		panic("internal: FromTable and FromSelect are both nil")
	}

	var wroteWhere bool
	if n.TimeWhere != nil && n.TimeWhere.Expr != "" {
		wroteWhere = true
		b.out.WriteString(" WHERE (")
		b.out.WriteString(n.TimeWhere.Expr)
		b.out.WriteString(")")
		b.args = append(b.args, n.TimeWhere.Args...)
	}
	if n.Where != nil && n.Where.Expr != "" {
		if wroteWhere {
			b.out.WriteString(" AND (")
		} else {
			b.out.WriteString(" WHERE ")
		}
		b.out.WriteString(n.Where.Expr)
		if wroteWhere {
			b.out.WriteString(")")
		}
		b.args = append(b.args, n.Where.Args...)
	}

	if n.Group && len(n.DimFields) > 0 {
		b.out.WriteString(" GROUP BY ")
		for i := range n.DimFields {
			if i > 0 {
				b.out.WriteString(", ")
			}
			b.out.WriteString(strconv.Itoa(i + 1))
		}
	}

	if n.Having != nil && n.Having.Expr != "" {
		b.out.WriteString(" HAVING ")
		b.out.WriteString(n.Having.Expr)
		b.args = append(b.args, n.Having.Args...)
	}

	if len(n.OrderBy) > 0 {
		b.out.WriteString(" ORDER BY ")
		for i, f := range n.OrderBy {
			if i > 0 {
				b.out.WriteString(", ")
			}
			b.out.WriteString(b.ast.dialect.OrderByExpression(f.Name, f.Desc))
		}
	}

	if n.Limit != nil {
		b.out.WriteString(" LIMIT ")
		b.out.WriteString(strconv.FormatInt(*n.Limit, 10))
	}

	if n.Offset != nil {
		b.out.WriteString(" OFFSET ")
		b.out.WriteString(strconv.FormatInt(*n.Offset, 10))
	}

	return nil
}

func (b *sqlBuilder) writeJoin(joinType JoinType, baseSelect, joinSelect *SelectNode) error {
	b.out.WriteByte(' ')
	b.out.WriteString(string(joinType))
	b.out.WriteString(" JOIN ")
	// If the join select is a CTE, then just add the CTE alias otherwise add the full select query
	if !joinSelect.IsCTE {
		b.out.WriteByte('(')
		err := b.writeSelect(joinSelect)
		if err != nil {
			return err
		}
		b.out.WriteString(") ")
	}
	b.out.WriteString(joinSelect.Alias)

	if len(baseSelect.DimFields) == 0 {
		b.out.WriteString(" ON TRUE")
		return nil
	}

	b.out.WriteString(" ON ")
	for i, f := range baseSelect.DimFields {
		if i > 0 {
			b.out.WriteString(" AND ")
		}
		lhs := b.ast.sqlForMember(baseSelect.Alias, f.Name)
		rhs := b.ast.sqlForMember(joinSelect.Alias, f.Name)
		b.out.WriteByte('(')
		b.out.WriteString(b.ast.dialect.JoinOnExpression(lhs, rhs))
		b.out.WriteByte(')')
	}
	return nil
}
