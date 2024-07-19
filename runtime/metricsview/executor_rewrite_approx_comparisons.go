package metricsview

import "fmt"

// rewriteApproxComparisons rewrites the AST to use a LEFT or RIGHT join instead of a FULL joins for comparisons,
// which enables more efficient query execution at the cost of some accuracy.
// ---- CTE Optimization ---- //
// Extracts out the base or comparison query into a CTE depending on the sort field.
// This is done to enable more efficient query execution by adding filter in the join query to select only dimension values present in the CTE.
func (e *Executor) rewriteApproxComparisons(ast *AST) {
	if !e.instanceCfg.MetricsApproximateComparisons {
		return
	}

	_ = e.rewriteApproxComparisonsWalk(ast, ast.Root)
}

func (e *Executor) rewriteApproxComparisonsWalk(a *AST, n *SelectNode) bool {
	// If n is a comparison node, rewrite it
	var rewrote bool
	if n.JoinComparisonSelect != nil {
		rewrote = e.rewriteApproxComparisonNode(a, n)
	}

	// Recursively walk the base select.
	// NOTE: Probably doesn't matter, but should we walk the left join and comparison sub-selects?
	if n.FromSelect != nil {
		rewroteNested := e.rewriteApproxComparisonsWalk(a, n.FromSelect)
		rewrote = rewrote || rewroteNested
	}

	// If any node was rewritten, all parent nodes need to clear their offset (since it must only be applied once).
	if rewrote {
		n.Offset = nil
	}

	return rewrote
}

func (e *Executor) rewriteApproxComparisonNode(a *AST, n *SelectNode) bool {
	// Can only rewrite when sorting by exactly one field.
	if len(a.Root.OrderBy) != 1 {
		return false
	}
	sortField := a.Root.OrderBy[0]

	// Find out what we're sorting by
	var sortDim, sortBase, sortComparison, sortDelta bool
	var sortUnderlyingMeasure string
	if len(a.Root.OrderBy) > 0 {
		// Check if it's a measure
		for _, qm := range a.query.Measures {
			if qm.Name != sortField.Name {
				continue
			}

			if qm.Compute != nil && qm.Compute.ComparisonValue != nil {
				sortComparison = true
				sortUnderlyingMeasure = qm.Compute.ComparisonValue.Measure
			} else if qm.Compute != nil && qm.Compute.ComparisonDelta != nil {
				sortDelta = true
				sortUnderlyingMeasure = qm.Compute.ComparisonDelta.Measure
			} else if qm.Compute != nil && qm.Compute.ComparisonRatio != nil {
				sortDelta = true
				sortUnderlyingMeasure = qm.Compute.ComparisonRatio.Measure
			} else {
				sortBase = true
				sortUnderlyingMeasure = qm.Name
			}

			break
		}

		if !sortBase && !sortComparison && !sortDelta {
			// It wasn't a measure. Check if it's a dimension.
			for _, qd := range a.query.Dimensions {
				if qd.Name == sortField.Name {
					sortDim = true
					break
				}
			}
		}
	}

	// If sorting by a computed measure, we need to use the underlying measure name when pushing the order into the sub-select.
	if sortUnderlyingMeasure != "" {
		sortField.Name = sortUnderlyingMeasure
	}
	order := []OrderFieldNode{sortField}

	// Note: All these cases are approximations in different ways.
	if sortBase {
		// We're sorting by a measure in FromSelect. We do a LEFT JOIN and push down the order/limit to it.
		// This should remain correct when the limit is lower than the number of rows in the base query.
		// The approximate part here is when the base query returns fewer rows than the limit, then dimension values that are only in the comparison query will be missing.
		n.JoinComparisonType = JoinTypeLeft
		n.FromSelect.OrderBy = order
		n.FromSelect.Limit = a.Root.Limit
		n.FromSelect.Offset = a.Root.Offset

		// ---- CTE Optimization ---- //
		// make FromSelect a CTE and set IsCTE flag on FromSelect
		cte := n.FromSelect
		n.CTEs = append(n.CTEs, cte)
		n.FromSelect.IsCTE = true

		// now change the JoinComparisonSelect WHERE clause to use selected dim values from CTE
		for _, dim := range cte.DimFields {
			var dimExpr string
			if dim.Expr != "" {
				dimExpr = dim.Expr
			} else {
				dimExpr = dim.Name
			}
			n.JoinComparisonSelect.Where = n.JoinComparisonSelect.Where.and(fmt.Sprintf("%[1]s IS NULL OR %[1]s IN (SELECT %[2]q.%[3]q FROM %[2]q)", dimExpr, cte.Alias, dim.Name), nil)
		}
	} else if sortComparison {
		// We're sorting by a measure in JoinComparisonSelect. We can do a RIGHT JOIN and push down the order/limit to it.
		// This should remain correct when the limit is lower than the number of rows in the comparison query.
		// The approximate part here is when the comparison query returns fewer rows than the limit, then dimension values that are only in the base query will be missing.
		n.JoinComparisonType = JoinTypeRight
		n.JoinComparisonSelect.OrderBy = order
		n.JoinComparisonSelect.Limit = a.Root.Limit
		n.JoinComparisonSelect.Offset = a.Root.Offset

		// ---- CTE Optimization ---- //
		// make JoinComparisonSelect a CTE and set IsCTE flag on JoinComparisonSelect
		cte := n.JoinComparisonSelect
		n.CTEs = append(n.CTEs, cte)
		n.JoinComparisonSelect.IsCTE = true

		// now change the FromSelect WHERE clause to use selected dim values from CTE
		for _, dim := range cte.DimFields {
			var dimExpr string
			if dim.Expr != "" {
				dimExpr = dim.Expr
			} else {
				dimExpr = dim.Name
			}
			n.FromSelect.Where = n.FromSelect.Where.and(fmt.Sprintf("%[1]s IS NULL OR %[1]s IN (SELECT %[2]q.%[3]q FROM %[2]q)", dimExpr, cte.Alias, dim.Name), nil)
		}
	} else if sortDim {
		// We're sorting by a dimension. We do a LEFT JOIN that only returns values present in the base query.
		// The approximate part here is that dimension values only present in the comparison query will be missing.
		n.JoinComparisonType = JoinTypeLeft
		n.FromSelect.OrderBy = order
		n.FromSelect.Limit = a.Root.Limit
		n.FromSelect.Offset = a.Root.Offset

		// ---- CTE Optimization ---- //
		// make FromSelect a CTE and set IsCTE flag on FromSelect
		cte := n.FromSelect
		n.CTEs = append(n.CTEs, cte)
		n.FromSelect.IsCTE = true

		// now change the JoinComparisonSelect WHERE clause to use selected dim values from CTE
		for _, dim := range cte.DimFields {
			var dimExpr string
			if dim.Expr != "" {
				dimExpr = dim.Expr
			} else {
				dimExpr = dim.Name
			}
			n.JoinComparisonSelect.Where = n.JoinComparisonSelect.Where.and(fmt.Sprintf("%[1]s IS NULL OR %[1]s IN (SELECT %[2]q.%[3]q FROM %[2]q)", dimExpr, cte.Alias, dim.Name), nil)
		}
	}
	// TODO: Good ideas for approx delta sorts?

	return true
}
