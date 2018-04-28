package base

import (
	"fmt"
	//"github.com/satori/go.uuid"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

var genSymCounter int = 0
func MakeGenSym (name string) string {
	gensym := fmt.Sprintf("%s_%d", name, genSymCounter)
	genSymCounter++
	
	return gensym
}

func MakeProgram () *CXProgram {
	//heap := make([]byte, 0)
	newContext := &CXProgram{
		Packages: make([]*CXPackage, 0),
		CallStack: MakeCallStack(0),
		//Heap: &heap,
		Steps: make([][]CXCall, 0)}
	return newContext
}

func MakeParameterCopy (param *CXParameter) *CXParameter {
	return &CXParameter{
		Name: param.Name,
		Typ: param.Typ,
	}
}

func MakeArgumentCopy (arg *CXArgument) *CXArgument {
	value := *arg.Value
	return &CXArgument{
		Typ: arg.Typ,
		Value: &value,
	}
}

func MakeExpressionCopy (expr *CXExpression, fn *CXFunction, mod *CXPackage, cxt *CXProgram) *CXExpression {
	argsCopy := make([]*CXArgument, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		argsCopy[i] = MakeArgumentCopy(arg)
	}
	return &CXExpression{
		Operator: expr.Operator,
		Arguments: argsCopy,
		OutputNames: expr.OutputNames,
		Line: expr.Line,
		Function: fn,
		Package: mod,
		Program: cxt,
	}
}

func MakeFunctionCopy (fn *CXFunction, mod *CXPackage, cxt *CXProgram) *CXFunction {
	newFn := &CXFunction{}
	inputsCopy := make([]*CXParameter, len(fn.Inputs))
	outputsCopy := make([]*CXParameter, len(fn.Outputs))
	exprsCopy := make([]*CXExpression, len(fn.Expressions))
	for i, inp := range fn.Inputs {
		inputsCopy[i] = MakeParameterCopy(inp)
	}
	for i, out := range fn.Outputs {
		outputsCopy[i] = MakeParameterCopy(out)
	}
	
	for i, expr := range fn.Expressions {
		exprsCopy[i] = MakeExpressionCopy(expr, newFn, mod, cxt)
	}
	
	newFn.Name = fn.Name
	newFn.Inputs = inputsCopy
	newFn.Outputs = outputsCopy
	
	// if fn.Output != nil {
	// 	newFn.Output = MakeParameterCopy(fn.Output)
	// }
	newFn.Expressions = exprsCopy
	if len(exprsCopy) > 0 {
		newFn.CurrentExpression = exprsCopy[len(exprsCopy) - 1]
	}
	newFn.Package = mod
	newFn.Program = cxt

	return newFn
}

func MakeFieldCopy (fld *CXField) *CXField {
	return &CXField{
		Name: fld.Name,
		Typ: fld.Typ,
	}
}

func MakeStructCopy (strct *CXStruct, mod *CXPackage, cxt *CXProgram) *CXStruct {
	fldsCopy := make([]*CXField, len(strct.Fields))
	for i, fld := range strct.Fields {
		fldsCopy[i] = MakeFieldCopy(fld)
	}
	return &CXStruct{
		Name: strct.Name,
		Fields: fldsCopy,
		Package: mod,
		Program: cxt,
	}
}

func MakeDefinitionCopy (def *CXDefinition, mod *CXPackage, cxt *CXProgram) *CXDefinition {
	valCopy := *def.Value
	return &CXDefinition{
		Name: def.Name,
		Typ: def.Typ,
		Value: &valCopy,
		Package: mod,
		Program: cxt,
	}
}

func MakeModuleCopy (mod *CXPackage, cxt *CXProgram) *CXPackage {
	newMod := &CXPackage{Program: cxt}
	fnsCopy := make([]*CXFunction, len(mod.Functions))
	strctsCopy := make([]*CXStruct, len(mod.Structs))
	defsCopy := make([]*CXDefinition, len(mod.Definitions))
	
	for k, fn := range mod.Functions {
		fnsCopy[k] = MakeFunctionCopy(fn, newMod, cxt)
	}
	for k, strct := range mod.Structs {
		strctsCopy[k] = MakeStructCopy(strct, newMod, cxt)
	}
	for k, def := range mod.Definitions {
		defsCopy[k] = MakeDefinitionCopy(def, newMod, cxt)
	}

	// Setting current function in copy
	for _, fn := range fnsCopy {
		if fn.Name == mod.CurrentFunction.Name {
			newMod.CurrentFunction = fn
		}
	}

	newMod.Name = mod.Name
	newMod.Imports = mod.Imports
	newMod.Functions = fnsCopy
	newMod.Structs = strctsCopy
	newMod.Definitions = defsCopy
	newMod.Program = cxt
	
	return newMod
}

func MakeCallCopy (call *CXCall, mod *CXPackage, cxt *CXProgram) CXCall {
	stateCopy := make([]*CXDefinition, len(call.State))
	for k, v := range call.State {
		stateCopy[k] = MakeDefinitionCopy(v, mod, cxt)
	}
	return CXCall{
		Operator: call.Operator,
		Line: call.Line,
		State: stateCopy,
		ReturnAddress: call.ReturnAddress,
		Package: mod,
		Program: cxt,
	}
}

func MakeCallStack (size int) []CXCall {
	// return &CXCallStack{
	// 	Calls: make([]*CXCall, size),
	// }

	return make([]CXCall, size)
}

func MakeProgramCopy (cxt *CXProgram, stepNumber int) *CXProgram {
	newContext := &CXProgram{}

	modsCopy := make([]*CXPackage, len(cxt.Packages))
	if stepNumber >= len(cxt.Steps) || stepNumber < 0 {
		stepNumber = len(cxt.Steps) - 1
	}
	
	for k, mod := range cxt.Packages {
		modsCopy[k] = MakeModuleCopy(mod, newContext)
	}

	// Setting current module in copy
	for _, mod := range modsCopy {
		if mod.Name == cxt.CurrentPackage.Name {
			newContext.CurrentPackage = mod
		}
	}

	newContext.Packages = modsCopy

	// Making imports copies
	for _, mod := range modsCopy {
		for impKey, _ := range mod.Imports {
			mod.Imports[impKey] =  modsCopy[impKey]
		}
	}

	// Making expressions/operators
	for _, mod := range modsCopy {
		for _, fn := range mod.Functions {
			for _, expr := range fn.Expressions {
				if op, err := newContext.GetFunction(expr.Operator.Name, expr.Package.Name); err == nil {
					expr.Operator = op
				}
			}
		}
	}

	if len(cxt.Steps) > 0 {
		reqStep := cxt.Steps[stepNumber]
		newStep := MakeCallStack(len(reqStep))
		
		var lastCall *CXCall
		for j, call := range reqStep {
			var callModule *CXPackage
			for _, mod := range modsCopy {
				if call.Package.Name == mod.Name {
					callModule = mod
				}
			}
			
			newCall := MakeCallCopy(&call, callModule, newContext)
			if callOp, err := newContext.GetFunction(call.Operator.Name, call.Operator.Package.Name); err == nil {
				newCall.Operator = callOp
			}
			newCall.ReturnAddress = lastCall
			lastCall = &newCall
			newStep[j] = newCall
		}
		
		newContext.CallStack = newStep
		newContext.Steps = nil
	}
	
	return newContext
}

func MakeModule (name string) *CXPackage {
	return &CXPackage{
		Name: name,
		Definitions: make([]*CXDefinition, 0, 10),
		Imports: make([]*CXPackage, 0),
		Functions: make([]*CXFunction, 0, 10),
		Structs: make([]*CXStruct, 0),
	}
}

func MakeDefinition (name string, value *[]byte, typ string) *CXDefinition {
	return &CXDefinition{
		Name: name,
		Typ: typ,
		Value: value,
	}
}

func MakeField (name string, typ string) *CXField {
	return &CXField{Name: name, Typ: typ}
}

func MakeFieldFromParameter (param *CXParameter) *CXField {
	return &CXField{Name: param.Name, Typ: param.Typ}
}

// Used only for native types
func MakeDefaultValue (typName string) *[]byte {
	var zeroVal []byte
	switch typName {
	case "byte": zeroVal = make([]byte, 1, 1)
	case "i64", "f64": zeroVal = make([]byte, 8, 8)
	default: zeroVal = make([]byte, 4, 4)
	}
	return &zeroVal
}

func MakeStruct (name string) *CXStruct {
	return &CXStruct{Name: name}
}

func MakeParameter (name string, typ string) *CXParameter {
	return &CXParameter{Name: name,
		Typ: typ}
}

func MakeExpression (op *CXFunction) *CXExpression {
	return &CXExpression{Operator: op}
}

func MakeArgument (value *[]byte, typ string) *CXArgument {
	return &CXArgument{
		Typ: typ,
		Value: value,
	}
}

func MakeFunction (name string) *CXFunction {
	return &CXFunction{Name: name}
}

func MakeValue (value string) *[]byte {
	byts := encoder.Serialize(value)
	return &byts
}

func MakeCall (op *CXFunction, state []*CXDefinition, ret *CXCall, mod *CXPackage, cxt *CXProgram) *CXCall {
	return &CXCall{
		Operator: op,
		Line: 0,
		State: state,
		ReturnAddress: ret,
		Package: mod,
		Program: cxt,}
}

func MakeAffordance (desc string, action func()) *CXAffordance {
	return &CXAffordance{
		Description: desc,
		Action: action,
	}
}

func MakeIdentityOpName (typeName string) string {
	switch typeName {
	case "str":
		return "str.id"
	case "bool":
		return "bool.id"
	case "byte":
		return "byte.id"
	case "i32":
		return "i32.id"
	case "i64":
		return "i64.id"
	case "f32":
		return "f32.id"
	case "f64":
		return "f64.id"
	case "[]bool":
		return "[]bool.id"
	case "[]byte":
		return "[]byte.id"
	case "[]str":
		return "[]str.id"
	case "[]i32":
		return "[]i32.id"
	case "[]i64":
		return "[]i64.id"
	case "[]f32":
		return "[]f32.id"
	case "[]f64":
		return "[]f64.id"
	default:
		return ""
	}
}
