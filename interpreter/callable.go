package interpreter

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/wwqdrh/tinyparser/ast"
)

const nativeFuncStringForm = "<native fn>"

type Callable interface {
	Clean()
	Arity() int
	Call(interpreter *Interpreter, arguments []interface{}) (result interface{}, err error)
	String() string
}

type Function struct {
	Declaration   *ast.FunctionStmt
	Closure       *Environment
	IsInitializer bool
}

func (f *Function) Arity() int {
	return len(f.Declaration.Params)
}

func (f *Function) Call(interpreter *Interpreter, arguments []interface{}) (result interface{}, err error) {
	env := NewChildEnvironment(f.Closure)
	for i, param := range f.Declaration.Params {
		env.Define(param.Lexeme, arguments[i])
	}

	defer func() {
		if f.IsInitializer {
			var badThis error
			result, badThis = f.Closure.GetAt(0, "this")
			if badThis != nil {
				panic(badThis)
			}
		}
	}()
	defer func() {
		panicReason := recover()
		if panicReason == nil {
			return
		}
		returnValue, ok := panicReason.(*returnPayload)
		if !ok {
			panic(panicReason)
		}

		result = returnValue.Value
	}()

	err = interpreter.executeBlock(f.Declaration.Body, env)
	return
}

func (f *Function) String() string {
	return fmt.Sprintf("<fn %s>", f.Declaration.Name.Lexeme)
}

func (f *Function) Bind(i *Instance) *Function {
	env := NewChildEnvironment(f.Closure)
	env.Define("this", i)
	return &Function{
		Declaration:   f.Declaration,
		Closure:       env,
		IsInitializer: f.IsInitializer,
	}
}

type returnPayload struct {
	Value interface{}
}

type nativeFuncClock struct{}

func (n nativeFuncClock) Arity() int {
	return 0
}

func (n nativeFuncClock) Call(interpreter *Interpreter, arguments []interface{}) (result interface{}, err error) {
	return float64(time.Now().Unix()), nil
}

func (n nativeFuncClock) String() string {
	return nativeFuncStringForm
}

type nativeFuncSleep struct{}

func (n nativeFuncSleep) Arity() int {
	return 1
}

func (n nativeFuncSleep) Call(interpreter *Interpreter, arguments []interface{}) (result interface{}, err error) {
	ms, ok := arguments[0].(float64)
	if !ok {
		err = fmt.Errorf("sleep() requires milliseconds in float, not %T", arguments[0])
		return
	}

	time.Sleep(time.Duration(ms) * time.Millisecond)
	return nil, nil
}

func (n nativeFuncSleep) String() string {
	return nativeFuncStringForm
}

type nativeFuncRandN struct{}

func (n nativeFuncRandN) Arity() int {
	return 1
}

// Call returns, as an int,
// a non-negative pseudo-random number in the half-open interval [0,n) from the default Source.
// It returns error if n <= 0.
func (n nativeFuncRandN) Call(interpreter *Interpreter, arguments []interface{}) (result interface{}, err error) {
	max, ok := arguments[0].(float64)
	if !ok {
		err = fmt.Errorf("randN() requires parameter in float, not %T", arguments[0])
		return
	}
	if max <= 0 {
		err = fmt.Errorf("randN()'s parameter must be positive, got %f", max)
		return
	}

	return float64(rand.Intn(int(max))), nil
}

func (n nativeFuncRandN) String() string {
	return nativeFuncStringForm
}
