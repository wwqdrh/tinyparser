package interpreter

import (
	"fmt"

	"github.com/wwqdrh/tinyparser/token"
)

type RuntimeError struct {
	Reason string
	Token  *token.Token
}

func (r RuntimeError) Error() string {
	if r.Token != nil && r.Token.Type > 0 {
		return fmt.Sprintf("operation %q at line %d: %s", r.Token.Type, r.Token.Line, r.Reason)
	}

	return r.Reason
}
