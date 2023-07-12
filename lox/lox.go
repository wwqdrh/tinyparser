package lox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/wwqdrh/tinyparser/resolver"

	"github.com/wwqdrh/tinyparser/interpreter"

	"github.com/wwqdrh/tinyparser/parser"

	"github.com/wwqdrh/tinyparser/scanner"
)

type Lox struct {
	interpreter *interpreter.Interpreter
	externalFun map[string]interpreter.Callable
}

func NewLox(stdout io.Writer) *Lox {
	return &Lox{
		interpreter: interpreter.NewInterpreter(stdout),
	}
}

func (l *Lox) WithExternalFun(funs map[string]interpreter.Callable) *Lox {
	l.externalFun = funs
	return l
}

func (l *Lox) RunFile(ctx context.Context, path string) (err error) {
	script, err := os.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("reading script file: %w", err)
		return
	}

	err = l.Run(ctx, script)
	if err != nil {
		err = fmt.Errorf("running script: %w", err)
		return
	}

	return
}

func (l *Lox) RunPrompt(ctx context.Context) (err error) {
	lineReader := bufio.NewScanner(os.Stdin)

	fmt.Printf("> ")
	for lineReader.Scan() {
		line := lineReader.Bytes()
		if len(line) == 0 {
			break
		}

		err = l.Run(ctx, line)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("> ")
	}

	err = lineReader.Err()
	if err != nil {
		err = fmt.Errorf("reading input: %w", err)
		return
	}

	return
}

// Run provided script.
// context is used to early stop interpretation on statement level.
//
// The provided script is read only, should not be modified.
func (l *Lox) Run(ctx context.Context, script []byte) (err error) {
	s := scanner.NewScanner(script)
	tokens, err := s.ScanTokens()
	if err != nil {
		err = fmt.Errorf("scaning tokens: %w", err)
		return
	}

	p := parser.NewParser(tokens)
	stmts, err := p.Parse()
	if err != nil {
		return
	}

	resolve := resolver.NewResolver(l.interpreter)
	err = resolve.ResolveStmts(stmts)
	if err != nil {
		err = fmt.Errorf("resolving statements: %w", err)
		return
	}

	for name, fn := range l.externalFun {
		l.interpreter.DefineEnv(name, fn)
	}

	err = l.interpreter.Interpret(ctx, stmts)
	if err != nil {
		return
	}

	return
}

func (l *Lox) ChangeStdoutTo(writer io.Writer) {
	l.interpreter.ChangeStdoutTo(writer)
}
