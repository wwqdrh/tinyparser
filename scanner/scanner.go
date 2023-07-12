package scanner

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/wwqdrh/tinyparser/token"
)

type Scanner struct {
	source []byte
	tokens []*token.Token

	start   int
	current int
	line    int
}

func NewScanner(source []byte) *Scanner {
	return &Scanner{
		source:  source,
		tokens:  make([]*token.Token, 0),
		start:   0,
		current: 0,
		line:    1,
	}
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) ScanTokens() (tokens []*token.Token, err error) {
	for !s.isAtEnd() {
		s.start = s.current
		err = s.scanToken()
		if err != nil {
			err = fmt.Errorf("scaning token: %w", err)
			return
		}
	}

	s.tokens = append(s.tokens, &token.Token{
		Type:    token.EOF,
		Lexeme:  "",
		Literal: nil,
		Line:    s.line,
	})

	tokens = s.tokens
	return
}

func (s *Scanner) scanToken() (err error) {
	var c = s.advance()
	switch c {
	case ' ', '\r', '\t':
		return
	case '\n':
		s.line++
		return
	case '(':
		s.addSimpleToken(token.LeftParen)
		return
	case ')':
		s.addSimpleToken(token.RightParen)
		return
	case '[':
		s.addSimpleToken(token.LeftSquare)
		return
	case ']':
		s.addSimpleToken(token.RightSquare)
		return
	case '{':
		s.addSimpleToken(token.LeftBrace)
		return
	case '}':
		s.addSimpleToken(token.RightBrace)
		return
	case ',':
		s.addSimpleToken(token.Comma)
		return
	case '.':
		s.addSimpleToken(token.Dot)
		return
	case '-':
		s.addSimpleToken(token.Minus)
		return
	case '+':
		s.addSimpleToken(token.Plus)
		return
	case ';':
		s.addSimpleToken(token.Semicolon)
		return
	case '*':
		s.addSimpleToken(token.Star)
		return
	case '!':
		if s.match('=') {
			s.addSimpleToken(token.BangEqual)
		} else {
			s.addSimpleToken(token.Bang)
		}
		return
	case '=':
		if s.match('=') {
			s.addSimpleToken(token.EqualEqual)
		} else {
			s.addSimpleToken(token.Equal)
		}
		return
	case '<':
		if s.match('=') {
			s.addSimpleToken(token.LessEqual)
		} else {
			s.addSimpleToken(token.Less)
		}
		return
	case '>':
		if s.match('=') {
			s.addSimpleToken(token.GreaterEqual)
		} else {
			s.addSimpleToken(token.Greater)
		}
		return
	case '/':
		// comments
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addSimpleToken(token.Slash)
		}
		return
	case '"':
		err = s.string()
		if err != nil {
			err = fmt.Errorf("scanning string at line %d: %w", s.line, err)
			return
		}
		return
	case '\'':
		err = s.singleString()
		if err != nil {
			err = fmt.Errorf("scanning string at line %d: %w", s.line, err)
			return
		}
		return
	}

	if isDigit(c) {
		err = s.number()
		if err != nil {
			err = fmt.Errorf("scanning number at line %d: %w", s.line, err)
			return
		}

		return
	}

	if isAlpha(c) {
		s.identifier()
		return
	}

	err = fmt.Errorf("unexpected character %q at line %d", c, s.line)
	return
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0
	}

	next, _ := utf8.DecodeRune(s.source[s.current:])
	return next
}

func (s *Scanner) advance() (c rune) {
	c, size := utf8.DecodeRune(s.source[s.current:])
	s.current += size
	return
}

func (s *Scanner) addToken(tokenType token.Type, literal interface{}) {
	var text = s.source[s.start:s.current]
	s.tokens = append(s.tokens, &token.Token{
		Type:    tokenType,
		Lexeme:  string(text),
		Literal: literal,
		Line:    s.line,
	})
}

func (s *Scanner) addSimpleToken(tokenType token.Type) {
	s.addToken(tokenType, nil)
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}

	current, size := utf8.DecodeRune(s.source[s.current:])
	if current != expected {
		return false
	}

	s.current += size
	return true
}

// string not like original Lox, we handle escaped char as well,
// which makes Lox Playground more interesting.
//
// \a   U+0007 alert or bell
// \b   U+0008 backspace
// \f   U+000C form feed
// \n   U+000A line feed or newline
// \r   U+000D carriage return
// \t   U+0009 horizontal tab
// \v   U+000B vertical tab
// \\   U+005C backslash
// \"   U+0022 double quote
func (s *Scanner) string() (err error) {
	var b strings.Builder

	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		c := s.advance()
		if c != '\\' {
			b.WriteRune(c)
			continue
		}
		switch s.peek() {
		case 'a':
			b.WriteRune('\a')
		case 'b':
			b.WriteRune('\b')
		case 'f':
			b.WriteRune('\f')
		case 'n':
			b.WriteRune('\n')
		case 'r':
			b.WriteRune('\r')
		case 't':
			b.WriteRune('\t')
		case 'v':
			b.WriteRune('\v')
		case '\\':
			b.WriteRune('\\')
		case '"':
			b.WriteRune('"')
		default:
			err = fmt.Errorf("unexpected escaped char '\\%s'", string(s.peek()))
			return
		}
		s.advance()
	}

	if s.isAtEnd() {
		err = errors.New("unterminated string")
		return
	}

	// skip the closing "
	s.advance()

	s.addToken(token.String, b.String())

	return
}

func (s *Scanner) singleString() (err error) {
	var b strings.Builder

	for s.peek() != '\'' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		c := s.advance()
		if c != '\\' {
			b.WriteRune(c)
			continue
		}
		switch s.peek() {
		case 'a':
			b.WriteRune('\a')
		case 'b':
			b.WriteRune('\b')
		case 'f':
			b.WriteRune('\f')
		case 'n':
			b.WriteRune('\n')
		case 'r':
			b.WriteRune('\r')
		case 't':
			b.WriteRune('\t')
		case 'v':
			b.WriteRune('\v')
		case '\\':
			b.WriteRune('\\')
		case '\'':
			b.WriteRune('\'')
		default:
			err = fmt.Errorf("unexpected escaped char '\\%s'", string(s.peek()))
			return
		}
		s.advance()
	}

	if s.isAtEnd() {
		err = errors.New("unterminated string")
		return
	}

	// skip the closing "
	s.advance()

	s.addToken(token.String, b.String())

	return
}

func (s *Scanner) number() (err error) {
	for isDigit(s.peek()) {
		s.advance()
	}

	// look for a fractional part
	if s.peek() == '.' && isDigit(s.peekNext()) {
		// consume the "."
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	text := s.source[s.start:s.current]
	value, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		err = fmt.Errorf("parsing float on %q: %w", text, err)
		return
	}

	s.addToken(token.Number, value)

	return
}

func (s *Scanner) identifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	tokenType, ok := token.KeywordMapping[string(s.source[s.start:s.current])]
	if !ok {
		tokenType = token.Identifier
	}

	s.addSimpleToken(tokenType)
}

func (s *Scanner) peekNext() rune {
	current, size := utf8.DecodeRune(s.source[s.current:])
	if current == utf8.RuneError {
		return 0
	}

	offset := s.current + size
	if offset >= len(s.source) {
		return 0
	}

	next, _ := utf8.DecodeRune(s.source[offset:])

	return next
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c rune) bool {
	return isAlpha(c) || isDigit(c)
}
