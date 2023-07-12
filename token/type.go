//go:generate stringer -type Type

package token

type Type int

const (
	SingleCharacterTokenStart Type = iota

	LeftParen
	RightParen
	LeftSquare
	RightSquare
	LeftBrace
	RightBrace
	Comma
	Dot
	Minus
	Plus
	Semicolon
	Slash
	Star

	SingleCharacterTokenEnd

	OneOrTwoCharacterTokenStart

	Bang
	BangEqual
	Equal
	EqualEqual
	Greater
	GreaterEqual
	Less
	LessEqual

	OneOrTwoCharacterTokenEnd

	LiteralStart

	Identifier
	String
	Number

	LiteralEnd

	KeywordStart

	And
	Class
	Else
	False
	Fun
	For
	If
	Nil
	Or
	Print
	Return
	Super
	This
	True
	Var
	While
	Require

	KeywordEnd

	EOF
)

var KeywordMapping = map[string]Type{
	"and":     And,
	"class":   Class,
	"else":    Else,
	"false":   False,
	"for":     For,
	"fun":     Fun,
	"if":      If,
	"nil":     Nil,
	"or":      Or,
	"print":   Print,
	"return":  Return,
	"super":   Super,
	"this":    This,
	"true":    True,
	"var":     Var,
	"while":   While,
	"require": Require,
}
