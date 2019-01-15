package evaluator

import (
	"fmt"
	"github.com/yuya373/monkey/ast"
	"github.com/yuya373/monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{
			Value: node.Value,
		}
	case *ast.Boolean:
		return evalBoolean(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		v := Eval(node.ReturnValue, env)
		if isError(v) {
			return v
		}
		return &object.ReturnValue{Value: v}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		env := object.CloneEnvironment(env)
		return &object.Function{Parameters: params, Body: body, Env: env}
	case *ast.CallExpression:
		fn := Eval(node.Function, env)
		if isError(fn) {
			return fn
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(fn, args)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	}

	return nil
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ &&
		index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError(
			"index operator not supported: %s",
			left.Type(),
		)
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arr := array.(*object.Array)
	idx := index.(*object.Integer).Value

	max := int64(len(arr.Elements) - 1)
	if idx < 0 || max < idx {
		return NULL
	}

	return arr.Elements[idx]
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch f := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(f, args)
		evaluated := Eval(f.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return f.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func unwrapReturnValue(obj object.Object) object.Object {
	if rValue, ok := obj.(*object.ReturnValue); ok {
		return rValue.Value
	}

	return obj
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	maxArgsIdx := len(args) - 1
	for i, param := range fn.Parameters {
		fn.Env.Delete(param.Value)

		if i <= maxArgsIdx {
			env.Set(param.Value, args[i])
		}
	}

	return env
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, exp := range exps {
		evaluated := Eval(exp, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if val, ok := builtins[node.Value]; ok {
		return val
	}

	return newError(
		"identifier not found: " + node.Value,
	)
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}

	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{
		Message: fmt.Sprintf(format, a...),
	}
}

func evalBlockStatement(
	block *ast.BlockStatement,
	env *object.Environment,
) object.Object {
	var result object.Object

	for _, stmt := range block.Statements {
		result = Eval(stmt, env)

		if result != nil {
			t := result.Type()
			if t == object.RETURN_VALUE_OBJ ||
				t == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalIfExpression(
	exp *ast.IfExpression,
	env *object.Environment,
) object.Object {
	cond := Eval(exp.Condition, env)
	if isError(cond) {
		return cond
	}

	if isTruthy(cond) {
		return Eval(exp.Consequence, env)
	} else if exp.Alternative != nil {
		return Eval(exp.Alternative, env)
	}

	return NULL
}

func isTruthy(obj object.Object) bool {
	if obj == NULL || obj == FALSE {
		return false
	}

	return true
}

func evalInfixExpression(op string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ &&
		right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(op, left, right)
	case left.Type() == object.STRING_OBJ &&
		right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(op, left, right)
	case op == "==":
		return evalBoolean(left == right)
	case op == "!=":
		return evalBoolean(left != right)
	case left.Type() != right.Type():
		return newError(
			"type mismatch: %s %s %s",
			left.Type(),
			op,
			right.Type(),
		)
	default:
		return newError(
			"unknown operator: %s %s %s",
			left.Type(),
			op,
			right.Type(),
		)
	}
}

func evalStringInfixExpression(
	op string,
	left, right object.Object,
) object.Object {
	lVal := left.(*object.String).Value
	rVal := right.(*object.String).Value

	switch op {
	case "+":
		return &object.String{Value: lVal + rVal}
	default:
		return newError(
			"unknown operator: %s %s %s",
			left.Type(),
			op,
			right.Type(),
		)
	}
}

func evalIntegerInfixExpression(op string, left, right object.Object) object.Object {
	lVal := left.(*object.Integer).Value
	rVal := right.(*object.Integer).Value

	switch op {
	case "+":
		return &object.Integer{Value: lVal + rVal}
	case "-":
		return &object.Integer{Value: lVal - rVal}
	case "*":
		return &object.Integer{Value: lVal * rVal}
	case "/":
		return &object.Integer{Value: lVal / rVal}
	case "<":
		return evalBoolean(lVal < rVal)
	case ">":
		return evalBoolean(lVal > rVal)
	case "==":
		return evalBoolean(lVal == rVal)
	case "!=":
		return evalBoolean(lVal != rVal)
	default:
		return newError(
			"unknown operator: %s %s %s",
			left.Type(),
			op,
			right.Type(),
		)
	}
}

func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError(
			"unknown operator: %s%s",
			op,
			right.Type(),
		)
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError(
			"unknown operator: -%s",
			right.Type(),
		)
	}

	v := right.(*object.Integer).Value
	return &object.Integer{Value: -v}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalBoolean(b bool) *object.Boolean {
	if b {
		return TRUE
	}

	return FALSE
}

func evalProgram(
	stmts []ast.Statement,
	env *object.Environment,
) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)

		switch r := result.(type) {
		case *object.ReturnValue:
			return r.Value
		case *object.Error:
			return r
		}
	}

	return result
}
