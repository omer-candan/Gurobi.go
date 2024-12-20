package gurobi

// #include <gurobi_passthrough.h>
import "C"
import (
	"errors"
	"fmt"
)

// Model ...
// Gurobi model object
type Model struct {
	AsGRBModel  *C.GRBmodel
	Env         Env
	Variables   []Var
	Constraints []Constr
}

/*
MakeUninitializedError
Description:

	This function simply returns a fixed error for when the model is not initialized.
*/
func (model *Model) MakeUninitializedError() error {
	return fmt.Errorf("The gurobi model was not yet initialized!")
}

/*
Check
Description:

	Checks that the model has been properly created.
*/
func (model *Model) Check() error {
	// Check to see if pointer is nil
	if model == nil {
		return model.MakeUninitializedError()
	}

	// Check on env component
	err := model.Env.Check()
	if err != nil {
		return err
	}

	// If no other checks were flagged, then return nil.
	return nil
}

/*
NewModel
Description:

	Creates a new model from a given environment.
*/
func NewModel(modelname string, env *Env) (*Model, error) {
	err := env.Check()
	if err != nil {
		return nil, env.MakeUninitializedError()
	}

	var model *C.GRBmodel
	errcode := C.GRBnewmodel(env.env, &model, C.CString(modelname), 0, nil, nil, nil, nil, nil)
	if errcode != 0 {
		return nil, env.MakeError(errcode)
	}

	newenv := C.GRBgetenv(model)
	if newenv == nil {
		return nil, errors.New("Failed retrieve the environment")
	}

	return &Model{AsGRBModel: model, Env: Env{newenv}}, nil
}

/*
LoadModel
Description:

	Loads a model from a file
*/
func LoadModel(modelPath string, env *Env) (*Model, error) {
	err := env.Check()
	if err != nil {
		return nil, env.MakeUninitializedError()
	}

	var model *C.GRBmodel
	errcode := C.GRBreadmodel(env.env, C.CString(modelPath), &model)
	if errcode != 0 {
		return nil, env.MakeError(errcode)
	}

	newenv := C.GRBgetenv(model)
	if newenv == nil {
		return nil, errors.New("Failed retrieve the environment")
	}

	return &Model{AsGRBModel: model, Env: Env{newenv}}, nil
}

// Free ...
// free the model
func (model *Model) Free() {
	if model == nil {
		return
	}
	C.GRBfreemodel(model.AsGRBModel)
}

/*
AddVar
Description:

	Create a variable to the model
	This includes inputs for:
	- obj = Linear coefficient applied to this variable in the objective function (i.e. objective =  ... + obj * newvar + ...)
	- lb = Lower Bound
	- ub = Upper Bound
	- name = Name of the variable
	- constrs = Array of constraints in which the variable participates.
	- columns = Numerical values associated with non-zero values for the new variable?
				(I think these are coefficients appearing in the linear constraint for this variable)

Links:

	Comes from the C API for Gurobi.
	Documentation for 9.0: https://www.gurobi.com/documentation/9.0/refman/c_addvar.html
*/
func (model *Model) AddVar(vtype int8, obj float64, lb float64, ub float64, name string, constrs []*Constr, columns []float64) (*Var, error) {
	err := model.Check()
	if err != nil {
		return nil, err
	}

	if len(constrs) != len(columns) {
		return nil, errors.New("either the length of constrs or columns are wrong")
	}

	ind := make([]int32, len(constrs))
	for i, c := range constrs {
		if c.Index < 0 {
			return nil, errors.New("Invalid index in constrs")
		}
		ind[i] = c.Index
	}

	pind := (*C.int)(nil)
	pval := (*C.double)(nil)
	if len(ind) > 0 {
		pind = (*C.int)(&ind[0])
		pval = (*C.double)(&columns[0])
	}

	errCode := C.GRBaddvar(model.AsGRBModel, C.int(len(constrs)), pind, pval, C.double(obj), C.double(lb), C.double(ub), C.char(vtype), C.CString(name))
	if errCode != 0 {
		return nil, model.MakeError(errCode)
	}

	if err := model.Update(); err != nil {
		return nil, err
	}

	model.Variables = append(model.Variables, Var{model, int32(len(model.Variables))})
	return &model.Variables[len(model.Variables)-1], nil
}

/*
AddVars
Description:

	Adds the list of variables defined by the input slices.
*/
func (model *Model) AddVars(vtypes []int8, objs []float64, lbs []float64, ubs []float64, names []string, constrs [][]*Constr, columns [][]float64) ([]*Var, error) {
	// Input Processing
	err := model.AddVars_InputChecking(vtypes, objs, lbs, ubs, names, constrs, columns)
	if err != nil {
		return nil, err
	}

	numnz := 0
	for _, constr := range constrs {
		numnz += len(constr)
	}

	beg := make([]int32, len(constrs))
	ind := make([]int32, numnz)
	val := make([]float64, numnz)
	k := 0
	for i := 0; i < len(constrs); i++ {
		if len(constrs[i]) != len(columns[i]) {
			return nil, errors.New("")
		}

		for j := 0; j < len(constrs[i]); j++ {
			idx := constrs[i][j].Index
			if idx < 0 {
				return nil, errors.New("")
			}
			ind[k+j] = idx
			val[k+j] = columns[i][j]
		}

		beg[i] = int32(k)
		k += len(constrs[i])
	}

	vnames := make([](*C.char), len(vtypes))
	for i, n := range names {
		vnames[i] = C.CString(n)
	}

	pbeg := (*C.int)(nil)
	pind := (*C.int)(nil)
	pval := (*C.double)(nil)
	if len(beg) > 0 {
		pbeg = (*C.int)(&beg[0])
		pind = (*C.int)(&ind[0])
		pval = (*C.double)(&val[0])
	}

	pobjs := (*C.double)(nil)
	plbs := (*C.double)(nil)
	pubs := (*C.double)(nil)
	pvtypes := (*C.char)(nil)
	pnames := (**C.char)(nil)
	if len(vtypes) > 0 {
		pobjs = (*C.double)(&objs[0])
		plbs = (*C.double)(&lbs[0])
		pubs = (*C.double)(&ubs[0])
		pvtypes = (*C.char)(&vtypes[0])
		pnames = (**C.char)(&vnames[0])
	}

	errCode := C.GRBaddvars(model.AsGRBModel, C.int(len(vtypes)), C.int(numnz), pbeg, pind, pval, pobjs, plbs, pubs, pvtypes, pnames)
	if errCode != 0 {
		return nil, model.MakeError(errCode)
	}

	if err := model.Update(); err != nil {
		return nil, err
	}

	//fmt.Printf("len(vtypes)=%v\n", len(vtypes))

	vars := make([]*Var, len(vtypes))
	xcols := len(model.Variables)
	for i := xcols; i < xcols+len(vtypes); i++ {
		model.Variables = append(model.Variables, Var{model, int32(i)})
		vars[i] = &model.Variables[len(model.Variables)-1]
	}
	return vars, nil
}

func (model *Model) AddVarsWithTypes(count int, vtype int8) ([]*Var, error) {
	vtypes := make([]int8, count)

	for i := 0; i < count; i++ {
		vtypes[i] = vtype
	}

	pbeg := (*C.int)(nil)
	pind := (*C.int)(nil)
	pval := (*C.double)(nil)

	pobjs := (*C.double)(nil)
	plbs := (*C.double)(nil)
	pubs := (*C.double)(nil)
	pvtypes := (*C.char)(nil)
	pnames := (**C.char)(nil)

	pvtypes = (*C.char)(&vtypes[0])

	errCode := C.GRBaddvars(model.AsGRBModel, C.int(count), C.int(0), pbeg, pind, pval, pobjs, plbs, pubs, pvtypes, pnames)
	if errCode != 0 {
		return nil, model.MakeError(errCode)
	}

	if err := model.Update(); err != nil {
		return nil, err
	}

	//fmt.Printf("len(vtypes)=%v\n", len(vtypes))

	vars := make([]*Var, len(vtypes))
	xcols := len(model.Variables)
	for i := xcols; i < xcols+len(vtypes); i++ {
		model.Variables = append(model.Variables, Var{model, int32(i)})
		vars[i-xcols] = &model.Variables[len(model.Variables)-1]
	}
	return vars, nil
}

func (model *Model) AddVarsWithoutTypes(lbs []float64, ubs []float64) ([]*Var, error) {

	pbeg := (*C.int)(nil)
	pind := (*C.int)(nil)
	pval := (*C.double)(nil)

	pobjs := (*C.double)(nil)
	plbs := (*C.double)(nil)
	pubs := (*C.double)(nil)
	pvtypes := (*C.char)(nil)
	pnames := (**C.char)(nil)

	plbs = (*C.double)(&lbs[0])
	pubs = (*C.double)(&ubs[0])

	errCode := C.GRBaddvars(model.AsGRBModel, C.int(len(lbs)), C.int(0), pbeg, pind, pval, pobjs, plbs, pubs, pvtypes, pnames)
	if errCode != 0 {
		return nil, model.MakeError(errCode)
	}

	if err := model.Update(); err != nil {
		return nil, err
	}

	//fmt.Printf("len(vtypes)=%v\n", len(vtypes))

	vars := make([]*Var, len(lbs))
	xcols := len(model.Variables)
	for i := xcols; i < xcols+len(lbs); i++ {
		model.Variables = append(model.Variables, Var{model, int32(i)})
		vars[i-xcols] = &model.Variables[len(model.Variables)-1]
	}
	return vars, nil
}

func (model *Model) AddVars_InputChecking(vtypes []int8, objs []float64, lbs []float64, ubs []float64, names []string, constrs [][]*Constr, columns [][]float64) error {
	// Check the model
	err := model.Check()
	if err != nil {
		return model.MakeUninitializedError()
	}

	// Check the length of each of the slices.
	if len(vtypes) != len(objs) {
		return MismatchedLengthError{
			Length1: len(vtypes),
			Length2: len(objs),
			Name1:   "vtypes",
			Name2:   "objs",
		}
	}

	if len(objs) != len(lbs) {
		return MismatchedLengthError{
			Length1: len(objs),
			Name1:   "objs",
			Length2: len(lbs),
			Name2:   "lbs",
		}
	}

	if len(lbs) != len(ubs) {
		return MismatchedLengthError{
			Length1: len(lbs),
			Name1:   "lbs",
			Length2: len(ubs),
			Name2:   "ubs",
		}
	}

	if len(ubs) != len(names) {
		return MismatchedLengthError{
			Length1: len(ubs),
			Name1:   "ubs",
			Length2: len(names),
			Name2:   "names",
		}
	}

	if len(constrs) > 0 {
		if len(names) != len(constrs) {
			return MismatchedLengthError{
				Length1: len(names),
				Name1:   "names",
				Length2: len(constrs),
				Name2:   "constrs",
			}
		}
	}

	if len(constrs) > 0 {
		if len(constrs) != len(columns) {
			return MismatchedLengthError{
				Length1: len(constrs),
				Name1:   "constrs",
				Length2: len(columns),
				Name2:   "columns",
			}
		}
	}

	// Everything is good!
	return nil
}

/*
AddConstr
Description:

	Add a Linear constraint into the model.
	Uses the GRBaddconstr() method from the C api.

Inputs:
  - vars: A slice of variable arrays which provide the indices for the gurobi model's variables.
  - val: A slice of float values which are used as coefficients for the variables in the linear constraint.
  - sense: A flag which determines if this is an equality, less than equal or greater than or equal constraint.
  - rhs: A float value which determines the constant which is on the other side of the constraint.
  - constrname: An optional name for the constraint.

Link:

	https://www.gurobi.com/documentation/9.1/refman/c_addconstr.html
*/
func (model *Model) AddConstr(vars []*Var, val []float64, sense int8, rhs float64, constrname string) (*Constr, error) {
	err := model.Check()
	if err != nil {
		return nil, model.MakeUninitializedError()
	}

	ind := make([]int32, len(vars))
	for i, v := range vars {
		if v.Index < 0 {
			return nil, errors.New("Invalid vars")
		}
		ind[i] = v.Index
	}

	pind := (*C.int)(nil)
	pval := (*C.double)(nil)
	if len(ind) > 0 {
		pind = (*C.int)(&ind[0])
		pval = (*C.double)(&val[0])
	}

	//fmt.Printf("pind = %v\n", *pind)
	//fmt.Printf("ind = %v\n vars[0] = %v\n", ind, vars[0].Index)

	var length int32
	length = (int32)(len(ind))

	C.GRBclean2((*C.int)(&length), pind, pval)

	errCode := C.GRBaddconstr(
		model.AsGRBModel,
		C.int(length),
		pind, pval,
		C.char(sense), C.double(rhs), C.CString(constrname))
	if errCode != 0 {
		return nil, model.MakeError(errCode)
	}

	if err := model.Update(); err != nil {
		return nil, err
	}

	model.Constraints = append(model.Constraints, Constr{model, int32(len(model.Constraints))})
	return &model.Constraints[len(model.Constraints)-1], nil
}

/*
AddConstrs
Description:

	Adds a set of constraints at once.
*/
func (model *Model) AddConstrs(vars [][]*Var, vals [][]float64, senses []int8, rhs []float64, constrnames []string) ([]*Constr, error) {
	err := model.Check()
	if err != nil {
		return nil, model.MakeUninitializedError()
	}

	err = model.InputChecking_AddConstrs(vars, vals, senses, rhs, constrnames)
	if err != nil {
		return nil, err
	}

	numnz := 0
	for _, v := range vars {
		numnz += len(v)
	}

	beg := make([]int32, len(constrnames))
	ind := make([]int32, numnz)
	_vals := make([]float64, numnz)
	k := 0
	for i := 0; i < len(vars); i++ {
		if len(vars[i]) != len(vals[i]) {
			return nil, errors.New("")
		}

		for j := 0; j < len(vars[i]); j++ {
			idx := vars[i][j].Index
			if idx < 0 {
				return nil, errors.New("")
			}
			ind[k+j] = idx
			_vals[k+j] = vals[i][j]
		}

		beg[i] = int32(k)
		k += len(vars[i])
	}

	name := make([](*C.char), len(constrnames))
	for i, n := range constrnames {
		name[i] = C.CString(n)
	}

	pbeg := (*C.int)(nil)
	pind := (*C.int)(nil)
	pvals := (*C.double)(nil)
	if len(beg) > 0 {
		pbeg = (*C.int)(&beg[0])
		pind = (*C.int)(&ind[0])
		pvals = (*C.double)(&_vals[0])
	}

	psenses := (*C.char)(nil)
	prhs := (*C.double)(nil)
	pname := (**C.char)(nil)
	if len(constrnames) > 0 {
		psenses = (*C.char)(&senses[0])
		prhs = (*C.double)(&rhs[0])
		pname = (**C.char)(&name[0])
	}

	errCode := C.GRBaddconstrs(model.AsGRBModel, C.int(len(constrnames)), C.int(numnz), pbeg, pind, pvals, psenses, prhs, pname)
	if errCode != 0 {
		return nil, model.MakeError(errCode)
	}

	if err := model.Update(); err != nil {
		return nil, err
	}

	constrs := make([]*Constr, len(constrnames))
	xrows := len(model.Constraints)
	for i := xrows; i < xrows+len(constrnames); i++ {
		model.Constraints = append(model.Constraints, Constr{model, int32(i)})
		constrs[i] = &model.Constraints[len(model.Constraints)-1]
	}
	return constrs, nil
}

/*
InputChecking_AddConstrs
Description:

	Checks the inputs to the AddConstrs function.
*/
func (model *Model) InputChecking_AddConstrs(vars [][]*Var, vals [][]float64, senses []int8, rhs []float64, constrnames []string) error {
	// Check the model
	err := model.Check()
	if err != nil {
		return model.MakeUninitializedError()
	}

	// Check the length of each of the slices.
	if len(vars) != len(vals) {
		return MismatchedLengthError{
			Length1: len(vars),
			Length2: len(vals),
			Name1:   "vars",
			Name2:   "vals",
		}
	}

	if len(vals) != len(senses) {
		return MismatchedLengthError{
			Length1: len(vals),
			Name1:   "vals",
			Length2: len(senses),
			Name2:   "senses",
		}
	}

	if len(senses) != len(rhs) {
		return MismatchedLengthError{
			Length1: len(senses),
			Name1:   "senses",
			Length2: len(rhs),
			Name2:   "rhs",
		}
	}

	if len(rhs) != len(constrnames) {
		return MismatchedLengthError{
			Length1: len(rhs),
			Name1:   "rhs",
			Length2: len(constrnames),
			Name2:   "constrnames",
		}
	}

	//if len(constrs) > 0 {
	//	if len(names) != len(constrs) {
	//		return MismatchedLengthError{
	//			Length1: len(names),
	//			Name1:   "names",
	//			Length2: len(constrs),
	//			Name2:   "constrs",
	//		}
	//	}
	//}
	//
	//if len(constrs) > 0 {
	//	if len(constrs) != len(columns) {
	//		return MismatchedLengthError{
	//			Length1: len(constrs),
	//			Name1:   "constrs",
	//			Length2: len(columns),
	//			Name2:   "columns",
	//		}
	//	}
	//}

	// Everything is good!
	return nil
}

// SetObjective ...
func (model *Model) SetObjective(objectiveExpr interface{}, sense int32) error {

	// Clear Out All Previous Quadratic Objective Terms
	if err := C.GRBdelq(model.AsGRBModel); err != 0 {
		return model.MakeError(err)
	}

	// Detect the Type of Objective We Have
	switch objectiveExpr.(type) {
	case *LinExpr:
		le := objectiveExpr.(*LinExpr)
		if err := model.SetLinearObjective(le, sense); err != nil {
			return err
		}
	case *QuadExpr:
		qe := objectiveExpr.(*QuadExpr)
		if err := model.SetQuadraticObjective(qe, sense); err != nil {
			return err
		}
	default:
		return errors.New("Unexpected objective expression type!")
	}

	return nil
}

/*
SetLinearObjective
Description:

	Adds a linear objective to the model.
*/
func (model *Model) SetLinearObjective(expr *LinExpr, sense int32) error {
	// Constants

	// Algorithm
	for tempIndex, tempVar := range expr.Ind {
		// Add Each to the objective, by modifying the obj attribute
		if err := tempVar.SetObj(expr.Val[tempIndex]); err != nil {
			return err
		}
	}

	// if err := model.SetDoubleAttrVars(C.GRB_DBL_ATTR_OBJ, expr.lind, expr.lval); err != nil {
	// 	return err
	// }
	if err := model.SetDoubleAttr(C.GRB_DBL_ATTR_OBJCON, expr.Offset); err != nil {
		return err
	}
	if err := model.SetIntAttr(C.GRB_INT_ATTR_MODELSENSE, sense); err != nil {
		return err
	}

	// If you successfully complete all steps, then return no errors.
	return nil
}

/*
SetQuadraticObjective
Description:

	Adds a quadratic objective to the model.
*/
func (model *Model) SetQuadraticObjective(expr *QuadExpr, sense int32) error {
	// Constants

	// Algorithm
	if err := model.addQPTerms(expr.qrow, expr.qcol, expr.qval); err != nil {
		return err
	}

	if err := model.SetDoubleAttrVars(C.GRB_DBL_ATTR_OBJ, expr.lind, expr.lval); err != nil {
		return err
	}
	if err := model.SetDoubleAttr(C.GRB_DBL_ATTR_OBJCON, expr.offset); err != nil {
		return err
	}
	if err := model.SetIntAttr(C.GRB_INT_ATTR_MODELSENSE, sense); err != nil {
		return err
	}

	// If you successfully complete all steps, then return no errors.
	return nil
}

func (model *Model) addQPTerms(qrow []*Var, qcol []*Var, qval []float64) error {
	if model == nil {
		return errors.New("")
	}

	if len(qrow) != len(qcol) || len(qcol) != len(qval) {
		return errors.New("")
	}

	_qrow := make([]int32, len(qrow))
	_qcol := make([]int32, len(qcol))
	for i := 0; i < len(qrow); i++ {
		if qrow[i].Index < 0 {
			return errors.New("")
		}
		if qcol[i].Index < 0 {
			return errors.New("")
		}

		_qrow[i] = qrow[i].Index
		_qcol[i] = qcol[i].Index
	}

	pqrow := (*C.int)(nil)
	pqcol := (*C.int)(nil)
	pqval := (*C.double)(nil)
	if len(qrow) > 0 {
		pqrow = (*C.int)(&_qrow[0])
		pqcol = (*C.int)(&_qcol[0])
		pqval = (*C.double)(&qval[0])
	}

	err := C.GRBaddqpterms(model.AsGRBModel, C.int(len(qrow)), pqrow, pqcol, pqval)
	if err != 0 {
		return model.MakeError(err)
	}

	return nil
}

// Update ...
func (model *Model) Update() error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBupdatemodel(model.AsGRBModel)
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

// Optimize ...
func (model *Model) Optimize() error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBoptimize(model.AsGRBModel)
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

// Write ...
func (model *Model) Write(filename string) error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBwrite(model.AsGRBModel, C.CString(filename))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

func (model *Model) NumVars() (int32, error) {
	return model.GetIntAttr(INT_ATTR_NUMVARS)
}

func (model *Model) NumConstrs() (int32, error) {
	return model.GetIntAttr(INT_ATTR_NUMCONSTRS)
}

// GetIntAttr ...
func (model *Model) GetIntAttr(attrname string) (int32, error) {
	if model == nil {
		return 0, errors.New("")
	}
	var attr int32
	err := C.GRBgetintattr(model.AsGRBModel, C.CString(attrname), (*C.int)(&attr))
	if err != 0 {
		return 0, model.MakeError(err)
	}
	return attr, nil
}

// GetDoubleAttr ...
func (model *Model) GetDoubleAttr(attrname string) (float64, error) {
	if model == nil {
		return 0, errors.New("")
	}
	var attr float64
	err := C.GRBgetdblattr(model.AsGRBModel, C.CString(attrname), (*C.double)(&attr))
	if err != 0 {
		return 0, model.MakeError(err)
	}
	return attr, nil
}

// GetStringAttr ...
func (model *Model) GetStringAttr(attrname string) (string, error) {
	if model == nil {
		return "", errors.New("")
	}
	var attr *C.char
	err := C.GRBgetstrattr(model.AsGRBModel, C.CString(attrname), (**C.char)(&attr))
	if err != 0 {
		return "", model.MakeError(err)
	}
	return C.GoString(attr), nil
}

// SetIntAttr ...
func (model *Model) SetIntAttr(attrname string, value int32) error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBsetintattr(model.AsGRBModel, C.CString(attrname), C.int(value))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

// SetDoubleAttr ...
func (model *Model) SetDoubleAttr(attrname string, value float64) error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBsetdblattr(model.AsGRBModel, C.CString(attrname), C.double(value))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

// SetStringAttr ...
func (model *Model) SetStringAttr(attrname string, value string) error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBsetstrattr(model.AsGRBModel, C.CString(attrname), C.CString(value))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

// GetDoubleAttrVars ...
func (model *Model) GetDoubleAttrVars(attrname string, vars []*Var) ([]float64, error) {
	ind := make([]int32, len(vars))
	for i, v := range vars {
		if v.Index < 0 {
			return []float64{}, errors.New("")
		}
		ind[i] = v.Index
	}
	return model.getDoubleAttrList(attrname, ind)
}

// SetDoubleAttrVars ...
func (model *Model) SetDoubleAttrVars(attrname string, vars []*Var, value []float64) error {
	ind := make([]int32, len(vars))
	for i, v := range vars {
		if v.Index < 0 {
			return errors.New("")
		}
		ind[i] = v.Index
	}
	return model.setDoubleAttrList(attrname, ind, value)
}

func (model *Model) getIntAttrElement(attr string, ind int32) (int32, error) {
	if model == nil {
		return 0.0, model.MakeUninitializedError()
	}
	var value int32
	err := C.GRBgetintattrelement(model.AsGRBModel, C.CString(attr), C.int(ind), (*C.int)(&value))
	if err != 0 {
		return 0, model.MakeError(err)
	}
	return value, nil
}

func (model *Model) getCharAttrElement(attr string, ind int32) (int8, error) {
	if model == nil {
		return 0, errors.New("")
	}
	var value int8
	err := C.GRBgetcharattrelement(model.AsGRBModel, C.CString(attr), C.int(ind), (*C.char)(&value))
	if err != 0 {
		return 0, model.MakeError(err)
	}
	return value, nil
}

func (model *Model) getDoubleAttrElement(attr string, ind int32) (float64, error) {
	if model == nil {
		return 0, errors.New("")
	}
	var value float64
	err := C.GRBgetdblattrelement(model.AsGRBModel, C.CString(attr), C.int(ind), (*C.double)(&value))
	if err != 0 {
		return 0, model.MakeError(err)
	}
	return value, nil
}

func (model *Model) getStringAttrElement(attr string, ind int32) (string, error) {
	if model == nil {
		return "", errors.New("")
	}
	var value *C.char
	err := C.GRBgetstrattrelement(model.AsGRBModel, C.CString(attr), C.int(ind), (**C.char)(&value))
	if err != 0 {
		return "", model.MakeError(err)
	}
	return C.GoString(value), nil
}

func (model *Model) setIntAttrElement(attr string, ind int32, value int32) error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBsetintattrelement(model.AsGRBModel, C.CString(attr), C.int(ind), C.int(value))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

func (model *Model) setCharAttrElement(attr string, ind int32, value int8) error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBsetcharattrelement(model.AsGRBModel, C.CString(attr), C.int(ind), C.char(value))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

func (model *Model) setDoubleAttrElement(attr string, ind int32, value float64) error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBsetdblattrelement(model.AsGRBModel, C.CString(attr), C.int(ind), C.double(value))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

func (model *Model) setStringAttrElement(attr string, ind int32, value string) error {
	if model == nil {
		return errors.New("")
	}
	err := C.GRBsetstrattrelement(model.AsGRBModel, C.CString(attr), C.int(ind), C.CString(value))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

func (model *Model) getDoubleAttrList(attrname string, ind []int32) ([]float64, error) {
	if model == nil {
		return []float64{}, errors.New("")
	}
	if len(ind) == 0 {
		return []float64{}, nil
	}
	value := make([]float64, len(ind))
	err := C.GRBgetdblattrlist(model.AsGRBModel, C.CString(attrname), C.int(len(ind)), (*C.int)(&ind[0]), (*C.double)(&value[0]))
	if err != 0 {
		return []float64{}, model.MakeError(err)
	}
	return value, nil
}

func (model *Model) setDoubleAttrList(attrname string, ind []int32, value []float64) error {
	if model == nil {
		return errors.New("")
	}
	if len(ind) != len(value) {
		return errors.New("")
	}
	if len(ind) == 0 {
		return nil
	}
	err := C.GRBsetdblattrlist(model.AsGRBModel, C.CString(attrname), C.int(len(ind)), (*C.int)(&ind[0]), (*C.double)(&value[0]))
	if err != 0 {
		return model.MakeError(err)
	}
	return nil
}

/*
GetVarByName
Description:
	Collects the GRBVar which has the given gurobi variable by name.
*/
