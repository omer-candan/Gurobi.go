package gurobi

// #include <gurobi_passthrough.h>
import "C"
import (
	"fmt"
)

type Env struct {
	env *C.GRBenv
}

// create a new environment.
func NewEnv(logfilename string) (*Env, error) {
	var env *C.GRBenv = nil
	errcode := int(C.GRBloadenv(&env, C.CString(logfilename)))
	if errcode != 0 {
		errMsg, err := C.GRBgeterrormsg(env)
		if err != nil {
			return nil, fmt.Errorf("Cannot create environment. Error code: %v. Error when getting corresponding error message: %v", errcode, err)
		}
		return nil, fmt.Errorf("Cannot create environment. Error code: %d. Error message: %v", errcode, (errMsg))
	}

	return &Env{env}, nil
}

// free environment.
func (env *Env) Free() {
	if env != nil {
		C.GRBfreeenv(env.env)
	}
}

/*
SetTimeLimit
Description:

	This member function is meant to set the time limit of the environment that has
	been created in Gurobi.
*/
func (env *Env) SetTimeLimit(limitIn float64) error {
	// Constants
	paramName := "TimeLimit"

	// Input Checking
	err := env.Check()
	if err != nil {
		return env.MakeUninitializedError()
	}

	// Algorithm
	errcode := int(C.GRBsetdblparam(env.env, C.CString(paramName), C.double(limitIn)))
	if errcode != 0 {
		return fmt.Errorf("There was an error running GRBsetdblparam(): Error code %v", errcode)
	}

	// If everything was successful, then return nil.
	return nil

}

/*
GetTimeLimit
Description:

	This member function is meant to set the time limit of the environment that has
	been created in Gurobi.
*/
func (env *Env) GetTimeLimit() (float64, error) {
	// Constants
	paramName := "TimeLimit"

	// Input Checking
	err := env.Check()
	if err != nil {
		return -1.0, env.MakeUninitializedError()
	}

	// Algorithm
	var limitOut C.double
	errcode := int(C.GRBgetdblparam(env.env, C.CString(paramName), &limitOut))
	if errcode != 0 {
		return -1, fmt.Errorf("There was an error running GRBsetdblparam(): Error code %v", errcode)
	}

	// If everything was successful, then return nil.
	return float64(limitOut), nil

}

/*
SetIntParam()
Description:

	Mirrors the functionality of the GRBsetintattr() function from the C api.
	Sets the parameter of the solver that has name paramName with value val.
*/
func (env *Env) SetIntParam(paramName string, val int32) error {
	// Check that the env object is not nil.
	if env == nil {
		return fmt.Errorf("env is nil!")
	}

	// Set Attribute
	errcode := int(C.GRBsetintparam(env.env, C.CString(paramName), C.int(val)))
	if errcode != 0 {
		return fmt.Errorf("There was an error running GRBsetintparam(), errcode %v", errcode)
	}

	// If everything was successful, then return nil.
	return nil
}

/*
SetDBLParam()
Description:

	Mirrors the functionality of the GRBsetdblattr() function from the C api.
	Sets the parameter of the solver that has name paramName with value val.
*/
func (env *Env) SetDBLParam(paramName string, val float64) error {
	// Check that attribute is actually a scalar double attribute.
	if !IsValidDBLParam(paramName) {
		return fmt.Errorf("The input attribute name (%v) is not considered a valid attribute.", paramName)
	}

	// Check that the env object is not nil.
	if env == nil {
		return fmt.Errorf("env is nil!")
	}

	// Set Attribute
	errcode := int(C.GRBsetdblparam(env.env, C.CString(paramName), C.double(val)))
	if errcode != 0 {
		return fmt.Errorf("There was an error running GRBsetdblparam(), errcode %v", errcode)
	}

	// If everything was successful, then return nil.
	return nil

}

/*
GetDBLParam()
Description:

	Mirrors the functionality of the GRBgetdblattr() function from the C api.
	Gets the parameter of the model with the name paramName if it exists.
*/
func (env *Env) GetDBLParam(paramName string) (float64, error) {
	// Check the paramName to make sure it is valid
	if !IsValidDBLParam(paramName) {
		return -1, fmt.Errorf("The input attribute name (%v) is not considered a valid attribute.", paramName)
	}

	// Check environment input
	if env == nil {
		return -1, fmt.Errorf("env is nil!")
	}

	// Use GRBgetdblparam
	var valOut C.double
	errcode := int(C.GRBgetdblparam(env.env, C.CString(paramName), &valOut))
	if errcode != 0 {
		return -1, fmt.Errorf("There was an error running GRBgetdblparam(). Errorcode %v", errcode)
	}

	// If everything was successful, then return nil.
	return float64(valOut), nil
}

func IsValidDBLParam(paramName string) bool {
	// All param names
	var scalarDoubleAttributes []string = []string{"TimeLimit", "Cutoff", "BestObjStop"}

	// Check that attribute is actually a scalar double attribute.
	paramNameIsValid := false

	for _, validName := range scalarDoubleAttributes {
		if validName == paramName {
			paramNameIsValid = true
			break
		}
	}

	return paramNameIsValid
}

/*
SetStringParam
Description:
*/
func (env *Env) SetStringParam(param string, newvalue string) error {
	err := env.Check()
	if err != nil {
		return env.MakeUninitializedError()
	}

	errcode := int(C.GRBsetstrparam(env.env, C.CString(param), C.CString(newvalue)))
	if errcode != 0 {
		return fmt.Errorf("There was an error running GRBsetstrparam(): Error code %v", errcode)
	}

	return nil
}

/*
Check
Description:

	Checks whether or not the given environment is well-formed.
*/
func (env *Env) Check() error {
	if env == nil {
		return env.MakeUninitializedError()
	}

	// Gurobi env (the sole member of gurobi.Env is not yet defined.
	if env.env == nil {
		return env.MakeUninitializedError()
	}

	// If all checks passed, return nil
	return nil
}
