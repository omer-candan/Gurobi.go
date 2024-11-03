package gurobi

// Linear expression of variables
type LinExpr struct {
	Ind    []*Var
	Val    []float64
	Offset float64
}

func (expr *LinExpr) AddTerm(v *Var, c float64) *LinExpr {
	expr.Ind = append(expr.Ind, v)
	expr.Val = append(expr.Val, c)
	return expr
}

func (expr *LinExpr) AddConstant(c float64) *LinExpr {
	expr.Offset += c
	return expr
}

// Linear expression of variables
type QuadExpr struct {
	lind   []*Var
	lval   []float64
	qrow   []*Var
	qcol   []*Var
	qval   []float64
	offset float64
}

func (expr *QuadExpr) AddTerm(v *Var, c float64) *QuadExpr {
	expr.lind = append(expr.lind, v)
	expr.lval = append(expr.lval, c)
	return expr
}

func (expr *QuadExpr) AddQTerm(v1 *Var, v2 *Var, c float64) *QuadExpr {
	expr.qrow = append(expr.qrow, v1)
	expr.qcol = append(expr.qcol, v2)
	expr.qval = append(expr.qval, c)
	return expr
}

func (expr *QuadExpr) AddConstant(c float64) *QuadExpr {
	expr.offset += c
	return expr
}
