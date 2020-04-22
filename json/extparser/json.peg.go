package extparser

// Code generated by peg json.peg DO NOT EDIT.

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleDocument
	ruleJSON
	ruleObject
	ruleArray
	ruleImport
	ruleString
	ruleSingleQuoteLiteral
	ruleDoubleQuoteLiteral
	ruleSingleQuoteEscape
	ruleDoubleQuoteEscape
	ruleUnicodeEscape
	ruleHexDigit
	ruleTrue
	ruleFalse
	ruleNull
	ruleNumber
	ruleMinus
	ruleIntegralPart
	ruleFractionalPart
	ruleExponentPart
	ruleSpacing
	ruleWhitespace
	ruleLongComment
	ruleLineComment
	rulePragma
	ruleLWING
	ruleRWING
	ruleLBRK
	ruleRBRK
	ruleCOMMA
	ruleCOLON
	ruleEOT
)

var rul3s = [...]string{
	"Unknown",
	"Document",
	"JSON",
	"Object",
	"Array",
	"Import",
	"String",
	"SingleQuoteLiteral",
	"DoubleQuoteLiteral",
	"SingleQuoteEscape",
	"DoubleQuoteEscape",
	"UnicodeEscape",
	"HexDigit",
	"True",
	"False",
	"Null",
	"Number",
	"Minus",
	"IntegralPart",
	"FractionalPart",
	"ExponentPart",
	"Spacing",
	"Whitespace",
	"LongComment",
	"LineComment",
	"Pragma",
	"LWING",
	"RWING",
	"LBRK",
	"RBRK",
	"COMMA",
	"COLON",
	"EOT",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(w io.Writer, pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Fprintf(w, " ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Fprintf(w, "%v %v\n", rule, quote)
			} else {
				fmt.Fprintf(w, "\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(w io.Writer, buffer string) {
	node.print(w, false, buffer)
}

func (node *node32) PrettyPrint(w io.Writer, buffer string) {
	node.print(w, true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(os.Stdout, buffer)
}

func (t *tokens32) WriteSyntaxTree(w io.Writer, buffer string) {
	t.AST().Print(w, buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(os.Stdout, buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	tree, i := t.tree, int(index)
	if i >= len(tree) {
		t.tree = append(tree, token32{pegRule: rule, begin: begin, end: end})
		return
	}
	tree[i] = token32{pegRule: rule, begin: begin, end: end}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type JSON struct {
	Buffer string
	buffer []rune
	rules  [33]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *JSON) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *JSON) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *JSON
	max token32
}

func (e *parseError) Error() string {
	tokens, err := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		err += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return err
}

func (p *JSON) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *JSON) WriteSyntaxTree(w io.Writer) {
	p.tokens32.WriteSyntaxTree(w, p.Buffer)
}

func Pretty(pretty bool) func(*JSON) error {
	return func(p *JSON) error {
		p.Pretty = pretty
		return nil
	}
}

func Size(size int) func(*JSON) error {
	return func(p *JSON) error {
		p.tokens32 = tokens32{tree: make([]token32, 0, size)}
		return nil
	}
}
func (p *JSON) Init(options ...func(*JSON) error) error {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	for _, option := range options {
		err := option(p)
		if err != nil {
			return err
		}
	}
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := p.tokens32
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Document <- <(Spacing JSON EOT)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleSpacing]() {
					goto l0
				}
				if !_rules[ruleJSON]() {
					goto l0
				}
				if !_rules[ruleEOT]() {
					goto l0
				}
				add(ruleDocument, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 JSON <- <((Object / Array / String / True / False / Null / Number / Import) Spacing)> */
		func() bool {
			position2, tokenIndex2 := position, tokenIndex
			{
				position3 := position
				{
					position4, tokenIndex4 := position, tokenIndex
					if !_rules[ruleObject]() {
						goto l5
					}
					goto l4
				l5:
					position, tokenIndex = position4, tokenIndex4
					if !_rules[ruleArray]() {
						goto l6
					}
					goto l4
				l6:
					position, tokenIndex = position4, tokenIndex4
					if !_rules[ruleString]() {
						goto l7
					}
					goto l4
				l7:
					position, tokenIndex = position4, tokenIndex4
					if !_rules[ruleTrue]() {
						goto l8
					}
					goto l4
				l8:
					position, tokenIndex = position4, tokenIndex4
					if !_rules[ruleFalse]() {
						goto l9
					}
					goto l4
				l9:
					position, tokenIndex = position4, tokenIndex4
					if !_rules[ruleNull]() {
						goto l10
					}
					goto l4
				l10:
					position, tokenIndex = position4, tokenIndex4
					if !_rules[ruleNumber]() {
						goto l11
					}
					goto l4
				l11:
					position, tokenIndex = position4, tokenIndex4
					if !_rules[ruleImport]() {
						goto l2
					}
				}
			l4:
				if !_rules[ruleSpacing]() {
					goto l2
				}
				add(ruleJSON, position3)
			}
			return true
		l2:
			position, tokenIndex = position2, tokenIndex2
			return false
		},
		/* 2 Object <- <(LWING (String COLON JSON COMMA)* (String COLON JSON)? RWING)> */
		func() bool {
			position12, tokenIndex12 := position, tokenIndex
			{
				position13 := position
				if !_rules[ruleLWING]() {
					goto l12
				}
			l14:
				{
					position15, tokenIndex15 := position, tokenIndex
					if !_rules[ruleString]() {
						goto l15
					}
					if !_rules[ruleCOLON]() {
						goto l15
					}
					if !_rules[ruleJSON]() {
						goto l15
					}
					if !_rules[ruleCOMMA]() {
						goto l15
					}
					goto l14
				l15:
					position, tokenIndex = position15, tokenIndex15
				}
				{
					position16, tokenIndex16 := position, tokenIndex
					if !_rules[ruleString]() {
						goto l16
					}
					if !_rules[ruleCOLON]() {
						goto l16
					}
					if !_rules[ruleJSON]() {
						goto l16
					}
					goto l17
				l16:
					position, tokenIndex = position16, tokenIndex16
				}
			l17:
				if !_rules[ruleRWING]() {
					goto l12
				}
				add(ruleObject, position13)
			}
			return true
		l12:
			position, tokenIndex = position12, tokenIndex12
			return false
		},
		/* 3 Array <- <(LBRK (JSON COMMA)* JSON? RBRK)> */
		func() bool {
			position18, tokenIndex18 := position, tokenIndex
			{
				position19 := position
				if !_rules[ruleLBRK]() {
					goto l18
				}
			l20:
				{
					position21, tokenIndex21 := position, tokenIndex
					if !_rules[ruleJSON]() {
						goto l21
					}
					if !_rules[ruleCOMMA]() {
						goto l21
					}
					goto l20
				l21:
					position, tokenIndex = position21, tokenIndex21
				}
				{
					position22, tokenIndex22 := position, tokenIndex
					if !_rules[ruleJSON]() {
						goto l22
					}
					goto l23
				l22:
					position, tokenIndex = position22, tokenIndex22
				}
			l23:
				if !_rules[ruleRBRK]() {
					goto l18
				}
				add(ruleArray, position19)
			}
			return true
		l18:
			position, tokenIndex = position18, tokenIndex18
			return false
		},
		/* 4 Import <- <('@' 'i' 'm' 'p' 'o' 'r' 't' '(' String ')')> */
		func() bool {
			position24, tokenIndex24 := position, tokenIndex
			{
				position25 := position
				if buffer[position] != rune('@') {
					goto l24
				}
				position++
				if buffer[position] != rune('i') {
					goto l24
				}
				position++
				if buffer[position] != rune('m') {
					goto l24
				}
				position++
				if buffer[position] != rune('p') {
					goto l24
				}
				position++
				if buffer[position] != rune('o') {
					goto l24
				}
				position++
				if buffer[position] != rune('r') {
					goto l24
				}
				position++
				if buffer[position] != rune('t') {
					goto l24
				}
				position++
				if buffer[position] != rune('(') {
					goto l24
				}
				position++
				if !_rules[ruleString]() {
					goto l24
				}
				if buffer[position] != rune(')') {
					goto l24
				}
				position++
				add(ruleImport, position25)
			}
			return true
		l24:
			position, tokenIndex = position24, tokenIndex24
			return false
		},
		/* 5 String <- <(SingleQuoteLiteral / DoubleQuoteLiteral)> */
		func() bool {
			position26, tokenIndex26 := position, tokenIndex
			{
				position27 := position
				{
					position28, tokenIndex28 := position, tokenIndex
					if !_rules[ruleSingleQuoteLiteral]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex = position28, tokenIndex28
					if !_rules[ruleDoubleQuoteLiteral]() {
						goto l26
					}
				}
			l28:
				add(ruleString, position27)
			}
			return true
		l26:
			position, tokenIndex = position26, tokenIndex26
			return false
		},
		/* 6 SingleQuoteLiteral <- <('\'' (SingleQuoteEscape / (!('\'' / '\\' / '\n' / '\r') .))* '\'')> */
		func() bool {
			position30, tokenIndex30 := position, tokenIndex
			{
				position31 := position
				if buffer[position] != rune('\'') {
					goto l30
				}
				position++
			l32:
				{
					position33, tokenIndex33 := position, tokenIndex
					{
						position34, tokenIndex34 := position, tokenIndex
						if !_rules[ruleSingleQuoteEscape]() {
							goto l35
						}
						goto l34
					l35:
						position, tokenIndex = position34, tokenIndex34
						{
							position36, tokenIndex36 := position, tokenIndex
							{
								position37, tokenIndex37 := position, tokenIndex
								if buffer[position] != rune('\'') {
									goto l38
								}
								position++
								goto l37
							l38:
								position, tokenIndex = position37, tokenIndex37
								if buffer[position] != rune('\\') {
									goto l39
								}
								position++
								goto l37
							l39:
								position, tokenIndex = position37, tokenIndex37
								if buffer[position] != rune('\n') {
									goto l40
								}
								position++
								goto l37
							l40:
								position, tokenIndex = position37, tokenIndex37
								if buffer[position] != rune('\r') {
									goto l36
								}
								position++
							}
						l37:
							goto l33
						l36:
							position, tokenIndex = position36, tokenIndex36
						}
						if !matchDot() {
							goto l33
						}
					}
				l34:
					goto l32
				l33:
					position, tokenIndex = position33, tokenIndex33
				}
				if buffer[position] != rune('\'') {
					goto l30
				}
				position++
				add(ruleSingleQuoteLiteral, position31)
			}
			return true
		l30:
			position, tokenIndex = position30, tokenIndex30
			return false
		},
		/* 7 DoubleQuoteLiteral <- <('"' (DoubleQuoteEscape / (!('"' / '\\' / '\n' / '\r') .))* '"')> */
		func() bool {
			position41, tokenIndex41 := position, tokenIndex
			{
				position42 := position
				if buffer[position] != rune('"') {
					goto l41
				}
				position++
			l43:
				{
					position44, tokenIndex44 := position, tokenIndex
					{
						position45, tokenIndex45 := position, tokenIndex
						if !_rules[ruleDoubleQuoteEscape]() {
							goto l46
						}
						goto l45
					l46:
						position, tokenIndex = position45, tokenIndex45
						{
							position47, tokenIndex47 := position, tokenIndex
							{
								position48, tokenIndex48 := position, tokenIndex
								if buffer[position] != rune('"') {
									goto l49
								}
								position++
								goto l48
							l49:
								position, tokenIndex = position48, tokenIndex48
								if buffer[position] != rune('\\') {
									goto l50
								}
								position++
								goto l48
							l50:
								position, tokenIndex = position48, tokenIndex48
								if buffer[position] != rune('\n') {
									goto l51
								}
								position++
								goto l48
							l51:
								position, tokenIndex = position48, tokenIndex48
								if buffer[position] != rune('\r') {
									goto l47
								}
								position++
							}
						l48:
							goto l44
						l47:
							position, tokenIndex = position47, tokenIndex47
						}
						if !matchDot() {
							goto l44
						}
					}
				l45:
					goto l43
				l44:
					position, tokenIndex = position44, tokenIndex44
				}
				if buffer[position] != rune('"') {
					goto l41
				}
				position++
				add(ruleDoubleQuoteLiteral, position42)
			}
			return true
		l41:
			position, tokenIndex = position41, tokenIndex41
			return false
		},
		/* 8 SingleQuoteEscape <- <('\\' ('b' / 't' / 'n' / 'f' / 'r' / '\'' / '\\' / '/' / UnicodeEscape))> */
		func() bool {
			position52, tokenIndex52 := position, tokenIndex
			{
				position53 := position
				if buffer[position] != rune('\\') {
					goto l52
				}
				position++
				{
					position54, tokenIndex54 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l55
					}
					position++
					goto l54
				l55:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('t') {
						goto l56
					}
					position++
					goto l54
				l56:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('n') {
						goto l57
					}
					position++
					goto l54
				l57:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('f') {
						goto l58
					}
					position++
					goto l54
				l58:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('r') {
						goto l59
					}
					position++
					goto l54
				l59:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('\'') {
						goto l60
					}
					position++
					goto l54
				l60:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('\\') {
						goto l61
					}
					position++
					goto l54
				l61:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('/') {
						goto l62
					}
					position++
					goto l54
				l62:
					position, tokenIndex = position54, tokenIndex54
					if !_rules[ruleUnicodeEscape]() {
						goto l52
					}
				}
			l54:
				add(ruleSingleQuoteEscape, position53)
			}
			return true
		l52:
			position, tokenIndex = position52, tokenIndex52
			return false
		},
		/* 9 DoubleQuoteEscape <- <('\\' ('b' / 't' / 'n' / 'f' / 'r' / '"' / '\\' / '/' / UnicodeEscape))> */
		func() bool {
			position63, tokenIndex63 := position, tokenIndex
			{
				position64 := position
				if buffer[position] != rune('\\') {
					goto l63
				}
				position++
				{
					position65, tokenIndex65 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l66
					}
					position++
					goto l65
				l66:
					position, tokenIndex = position65, tokenIndex65
					if buffer[position] != rune('t') {
						goto l67
					}
					position++
					goto l65
				l67:
					position, tokenIndex = position65, tokenIndex65
					if buffer[position] != rune('n') {
						goto l68
					}
					position++
					goto l65
				l68:
					position, tokenIndex = position65, tokenIndex65
					if buffer[position] != rune('f') {
						goto l69
					}
					position++
					goto l65
				l69:
					position, tokenIndex = position65, tokenIndex65
					if buffer[position] != rune('r') {
						goto l70
					}
					position++
					goto l65
				l70:
					position, tokenIndex = position65, tokenIndex65
					if buffer[position] != rune('"') {
						goto l71
					}
					position++
					goto l65
				l71:
					position, tokenIndex = position65, tokenIndex65
					if buffer[position] != rune('\\') {
						goto l72
					}
					position++
					goto l65
				l72:
					position, tokenIndex = position65, tokenIndex65
					if buffer[position] != rune('/') {
						goto l73
					}
					position++
					goto l65
				l73:
					position, tokenIndex = position65, tokenIndex65
					if !_rules[ruleUnicodeEscape]() {
						goto l63
					}
				}
			l65:
				add(ruleDoubleQuoteEscape, position64)
			}
			return true
		l63:
			position, tokenIndex = position63, tokenIndex63
			return false
		},
		/* 10 UnicodeEscape <- <('u' HexDigit HexDigit HexDigit HexDigit)> */
		func() bool {
			position74, tokenIndex74 := position, tokenIndex
			{
				position75 := position
				if buffer[position] != rune('u') {
					goto l74
				}
				position++
				if !_rules[ruleHexDigit]() {
					goto l74
				}
				if !_rules[ruleHexDigit]() {
					goto l74
				}
				if !_rules[ruleHexDigit]() {
					goto l74
				}
				if !_rules[ruleHexDigit]() {
					goto l74
				}
				add(ruleUnicodeEscape, position75)
			}
			return true
		l74:
			position, tokenIndex = position74, tokenIndex74
			return false
		},
		/* 11 HexDigit <- <([a-f] / [A-F] / [0-9])> */
		func() bool {
			position76, tokenIndex76 := position, tokenIndex
			{
				position77 := position
				{
					position78, tokenIndex78 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('f') {
						goto l79
					}
					position++
					goto l78
				l79:
					position, tokenIndex = position78, tokenIndex78
					if c := buffer[position]; c < rune('A') || c > rune('F') {
						goto l80
					}
					position++
					goto l78
				l80:
					position, tokenIndex = position78, tokenIndex78
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l76
					}
					position++
				}
			l78:
				add(ruleHexDigit, position77)
			}
			return true
		l76:
			position, tokenIndex = position76, tokenIndex76
			return false
		},
		/* 12 True <- <(('t' 'r' 'u' 'e') / ('T' 'r' 'u' 'e'))> */
		func() bool {
			position81, tokenIndex81 := position, tokenIndex
			{
				position82 := position
				{
					position83, tokenIndex83 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l84
					}
					position++
					if buffer[position] != rune('r') {
						goto l84
					}
					position++
					if buffer[position] != rune('u') {
						goto l84
					}
					position++
					if buffer[position] != rune('e') {
						goto l84
					}
					position++
					goto l83
				l84:
					position, tokenIndex = position83, tokenIndex83
					if buffer[position] != rune('T') {
						goto l81
					}
					position++
					if buffer[position] != rune('r') {
						goto l81
					}
					position++
					if buffer[position] != rune('u') {
						goto l81
					}
					position++
					if buffer[position] != rune('e') {
						goto l81
					}
					position++
				}
			l83:
				add(ruleTrue, position82)
			}
			return true
		l81:
			position, tokenIndex = position81, tokenIndex81
			return false
		},
		/* 13 False <- <(('f' 'a' 'l' 's' 'e') / ('F' 'a' 'l' 's' 'e'))> */
		func() bool {
			position85, tokenIndex85 := position, tokenIndex
			{
				position86 := position
				{
					position87, tokenIndex87 := position, tokenIndex
					if buffer[position] != rune('f') {
						goto l88
					}
					position++
					if buffer[position] != rune('a') {
						goto l88
					}
					position++
					if buffer[position] != rune('l') {
						goto l88
					}
					position++
					if buffer[position] != rune('s') {
						goto l88
					}
					position++
					if buffer[position] != rune('e') {
						goto l88
					}
					position++
					goto l87
				l88:
					position, tokenIndex = position87, tokenIndex87
					if buffer[position] != rune('F') {
						goto l85
					}
					position++
					if buffer[position] != rune('a') {
						goto l85
					}
					position++
					if buffer[position] != rune('l') {
						goto l85
					}
					position++
					if buffer[position] != rune('s') {
						goto l85
					}
					position++
					if buffer[position] != rune('e') {
						goto l85
					}
					position++
				}
			l87:
				add(ruleFalse, position86)
			}
			return true
		l85:
			position, tokenIndex = position85, tokenIndex85
			return false
		},
		/* 14 Null <- <(('n' 'u' 'l' 'l') / ('N' 'o' 'n' 'e'))> */
		func() bool {
			position89, tokenIndex89 := position, tokenIndex
			{
				position90 := position
				{
					position91, tokenIndex91 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l92
					}
					position++
					if buffer[position] != rune('u') {
						goto l92
					}
					position++
					if buffer[position] != rune('l') {
						goto l92
					}
					position++
					if buffer[position] != rune('l') {
						goto l92
					}
					position++
					goto l91
				l92:
					position, tokenIndex = position91, tokenIndex91
					if buffer[position] != rune('N') {
						goto l89
					}
					position++
					if buffer[position] != rune('o') {
						goto l89
					}
					position++
					if buffer[position] != rune('n') {
						goto l89
					}
					position++
					if buffer[position] != rune('e') {
						goto l89
					}
					position++
				}
			l91:
				add(ruleNull, position90)
			}
			return true
		l89:
			position, tokenIndex = position89, tokenIndex89
			return false
		},
		/* 15 Number <- <(Minus? IntegralPart FractionalPart? ExponentPart?)> */
		func() bool {
			position93, tokenIndex93 := position, tokenIndex
			{
				position94 := position
				{
					position95, tokenIndex95 := position, tokenIndex
					if !_rules[ruleMinus]() {
						goto l95
					}
					goto l96
				l95:
					position, tokenIndex = position95, tokenIndex95
				}
			l96:
				if !_rules[ruleIntegralPart]() {
					goto l93
				}
				{
					position97, tokenIndex97 := position, tokenIndex
					if !_rules[ruleFractionalPart]() {
						goto l97
					}
					goto l98
				l97:
					position, tokenIndex = position97, tokenIndex97
				}
			l98:
				{
					position99, tokenIndex99 := position, tokenIndex
					if !_rules[ruleExponentPart]() {
						goto l99
					}
					goto l100
				l99:
					position, tokenIndex = position99, tokenIndex99
				}
			l100:
				add(ruleNumber, position94)
			}
			return true
		l93:
			position, tokenIndex = position93, tokenIndex93
			return false
		},
		/* 16 Minus <- <'-'> */
		func() bool {
			position101, tokenIndex101 := position, tokenIndex
			{
				position102 := position
				if buffer[position] != rune('-') {
					goto l101
				}
				position++
				add(ruleMinus, position102)
			}
			return true
		l101:
			position, tokenIndex = position101, tokenIndex101
			return false
		},
		/* 17 IntegralPart <- <('0' / ([1-9] [0-9]*))> */
		func() bool {
			position103, tokenIndex103 := position, tokenIndex
			{
				position104 := position
				{
					position105, tokenIndex105 := position, tokenIndex
					if buffer[position] != rune('0') {
						goto l106
					}
					position++
					goto l105
				l106:
					position, tokenIndex = position105, tokenIndex105
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l103
					}
					position++
				l107:
					{
						position108, tokenIndex108 := position, tokenIndex
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l108
						}
						position++
						goto l107
					l108:
						position, tokenIndex = position108, tokenIndex108
					}
				}
			l105:
				add(ruleIntegralPart, position104)
			}
			return true
		l103:
			position, tokenIndex = position103, tokenIndex103
			return false
		},
		/* 18 FractionalPart <- <('.' [0-9]+)> */
		func() bool {
			position109, tokenIndex109 := position, tokenIndex
			{
				position110 := position
				if buffer[position] != rune('.') {
					goto l109
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l109
				}
				position++
			l111:
				{
					position112, tokenIndex112 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l112
					}
					position++
					goto l111
				l112:
					position, tokenIndex = position112, tokenIndex112
				}
				add(ruleFractionalPart, position110)
			}
			return true
		l109:
			position, tokenIndex = position109, tokenIndex109
			return false
		},
		/* 19 ExponentPart <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		func() bool {
			position113, tokenIndex113 := position, tokenIndex
			{
				position114 := position
				{
					position115, tokenIndex115 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l116
					}
					position++
					goto l115
				l116:
					position, tokenIndex = position115, tokenIndex115
					if buffer[position] != rune('E') {
						goto l113
					}
					position++
				}
			l115:
				{
					position117, tokenIndex117 := position, tokenIndex
					{
						position119, tokenIndex119 := position, tokenIndex
						if buffer[position] != rune('+') {
							goto l120
						}
						position++
						goto l119
					l120:
						position, tokenIndex = position119, tokenIndex119
						if buffer[position] != rune('-') {
							goto l117
						}
						position++
					}
				l119:
					goto l118
				l117:
					position, tokenIndex = position117, tokenIndex117
				}
			l118:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l113
				}
				position++
			l121:
				{
					position122, tokenIndex122 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l122
					}
					position++
					goto l121
				l122:
					position, tokenIndex = position122, tokenIndex122
				}
				add(ruleExponentPart, position114)
			}
			return true
		l113:
			position, tokenIndex = position113, tokenIndex113
			return false
		},
		/* 20 Spacing <- <(Whitespace / LongComment / LineComment / Pragma)*> */
		func() bool {
			{
				position124 := position
			l125:
				{
					position126, tokenIndex126 := position, tokenIndex
					{
						position127, tokenIndex127 := position, tokenIndex
						if !_rules[ruleWhitespace]() {
							goto l128
						}
						goto l127
					l128:
						position, tokenIndex = position127, tokenIndex127
						if !_rules[ruleLongComment]() {
							goto l129
						}
						goto l127
					l129:
						position, tokenIndex = position127, tokenIndex127
						if !_rules[ruleLineComment]() {
							goto l130
						}
						goto l127
					l130:
						position, tokenIndex = position127, tokenIndex127
						if !_rules[rulePragma]() {
							goto l126
						}
					}
				l127:
					goto l125
				l126:
					position, tokenIndex = position126, tokenIndex126
				}
				add(ruleSpacing, position124)
			}
			return true
		},
		/* 21 Whitespace <- <(' ' / '\t' / '\r' / '\n')+> */
		func() bool {
			position131, tokenIndex131 := position, tokenIndex
			{
				position132 := position
				{
					position135, tokenIndex135 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l136
					}
					position++
					goto l135
				l136:
					position, tokenIndex = position135, tokenIndex135
					if buffer[position] != rune('\t') {
						goto l137
					}
					position++
					goto l135
				l137:
					position, tokenIndex = position135, tokenIndex135
					if buffer[position] != rune('\r') {
						goto l138
					}
					position++
					goto l135
				l138:
					position, tokenIndex = position135, tokenIndex135
					if buffer[position] != rune('\n') {
						goto l131
					}
					position++
				}
			l135:
			l133:
				{
					position134, tokenIndex134 := position, tokenIndex
					{
						position139, tokenIndex139 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l140
						}
						position++
						goto l139
					l140:
						position, tokenIndex = position139, tokenIndex139
						if buffer[position] != rune('\t') {
							goto l141
						}
						position++
						goto l139
					l141:
						position, tokenIndex = position139, tokenIndex139
						if buffer[position] != rune('\r') {
							goto l142
						}
						position++
						goto l139
					l142:
						position, tokenIndex = position139, tokenIndex139
						if buffer[position] != rune('\n') {
							goto l134
						}
						position++
					}
				l139:
					goto l133
				l134:
					position, tokenIndex = position134, tokenIndex134
				}
				add(ruleWhitespace, position132)
			}
			return true
		l131:
			position, tokenIndex = position131, tokenIndex131
			return false
		},
		/* 22 LongComment <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		func() bool {
			position143, tokenIndex143 := position, tokenIndex
			{
				position144 := position
				if buffer[position] != rune('/') {
					goto l143
				}
				position++
				if buffer[position] != rune('*') {
					goto l143
				}
				position++
			l145:
				{
					position146, tokenIndex146 := position, tokenIndex
					{
						position147, tokenIndex147 := position, tokenIndex
						if buffer[position] != rune('*') {
							goto l147
						}
						position++
						if buffer[position] != rune('/') {
							goto l147
						}
						position++
						goto l146
					l147:
						position, tokenIndex = position147, tokenIndex147
					}
					if !matchDot() {
						goto l146
					}
					goto l145
				l146:
					position, tokenIndex = position146, tokenIndex146
				}
				if buffer[position] != rune('*') {
					goto l143
				}
				position++
				if buffer[position] != rune('/') {
					goto l143
				}
				position++
				add(ruleLongComment, position144)
			}
			return true
		l143:
			position, tokenIndex = position143, tokenIndex143
			return false
		},
		/* 23 LineComment <- <('/' '/' (!('\r' / '\n') .)* ('\r' / '\n'))> */
		func() bool {
			position148, tokenIndex148 := position, tokenIndex
			{
				position149 := position
				if buffer[position] != rune('/') {
					goto l148
				}
				position++
				if buffer[position] != rune('/') {
					goto l148
				}
				position++
			l150:
				{
					position151, tokenIndex151 := position, tokenIndex
					{
						position152, tokenIndex152 := position, tokenIndex
						{
							position153, tokenIndex153 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l154
							}
							position++
							goto l153
						l154:
							position, tokenIndex = position153, tokenIndex153
							if buffer[position] != rune('\n') {
								goto l152
							}
							position++
						}
					l153:
						goto l151
					l152:
						position, tokenIndex = position152, tokenIndex152
					}
					if !matchDot() {
						goto l151
					}
					goto l150
				l151:
					position, tokenIndex = position151, tokenIndex151
				}
				{
					position155, tokenIndex155 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l156
					}
					position++
					goto l155
				l156:
					position, tokenIndex = position155, tokenIndex155
					if buffer[position] != rune('\n') {
						goto l148
					}
					position++
				}
			l155:
				add(ruleLineComment, position149)
			}
			return true
		l148:
			position, tokenIndex = position148, tokenIndex148
			return false
		},
		/* 24 Pragma <- <('#' (!('\r' / '\n') .)* ('\r' / '\n'))> */
		func() bool {
			position157, tokenIndex157 := position, tokenIndex
			{
				position158 := position
				if buffer[position] != rune('#') {
					goto l157
				}
				position++
			l159:
				{
					position160, tokenIndex160 := position, tokenIndex
					{
						position161, tokenIndex161 := position, tokenIndex
						{
							position162, tokenIndex162 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l163
							}
							position++
							goto l162
						l163:
							position, tokenIndex = position162, tokenIndex162
							if buffer[position] != rune('\n') {
								goto l161
							}
							position++
						}
					l162:
						goto l160
					l161:
						position, tokenIndex = position161, tokenIndex161
					}
					if !matchDot() {
						goto l160
					}
					goto l159
				l160:
					position, tokenIndex = position160, tokenIndex160
				}
				{
					position164, tokenIndex164 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l165
					}
					position++
					goto l164
				l165:
					position, tokenIndex = position164, tokenIndex164
					if buffer[position] != rune('\n') {
						goto l157
					}
					position++
				}
			l164:
				add(rulePragma, position158)
			}
			return true
		l157:
			position, tokenIndex = position157, tokenIndex157
			return false
		},
		/* 25 LWING <- <('{' Spacing)> */
		func() bool {
			position166, tokenIndex166 := position, tokenIndex
			{
				position167 := position
				if buffer[position] != rune('{') {
					goto l166
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l166
				}
				add(ruleLWING, position167)
			}
			return true
		l166:
			position, tokenIndex = position166, tokenIndex166
			return false
		},
		/* 26 RWING <- <('}' Spacing)> */
		func() bool {
			position168, tokenIndex168 := position, tokenIndex
			{
				position169 := position
				if buffer[position] != rune('}') {
					goto l168
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l168
				}
				add(ruleRWING, position169)
			}
			return true
		l168:
			position, tokenIndex = position168, tokenIndex168
			return false
		},
		/* 27 LBRK <- <('[' Spacing)> */
		func() bool {
			position170, tokenIndex170 := position, tokenIndex
			{
				position171 := position
				if buffer[position] != rune('[') {
					goto l170
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l170
				}
				add(ruleLBRK, position171)
			}
			return true
		l170:
			position, tokenIndex = position170, tokenIndex170
			return false
		},
		/* 28 RBRK <- <(']' Spacing)> */
		func() bool {
			position172, tokenIndex172 := position, tokenIndex
			{
				position173 := position
				if buffer[position] != rune(']') {
					goto l172
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l172
				}
				add(ruleRBRK, position173)
			}
			return true
		l172:
			position, tokenIndex = position172, tokenIndex172
			return false
		},
		/* 29 COMMA <- <(',' Spacing)> */
		func() bool {
			position174, tokenIndex174 := position, tokenIndex
			{
				position175 := position
				if buffer[position] != rune(',') {
					goto l174
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l174
				}
				add(ruleCOMMA, position175)
			}
			return true
		l174:
			position, tokenIndex = position174, tokenIndex174
			return false
		},
		/* 30 COLON <- <(':' Spacing)> */
		func() bool {
			position176, tokenIndex176 := position, tokenIndex
			{
				position177 := position
				if buffer[position] != rune(':') {
					goto l176
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l176
				}
				add(ruleCOLON, position177)
			}
			return true
		l176:
			position, tokenIndex = position176, tokenIndex176
			return false
		},
		/* 31 EOT <- <!.> */
		func() bool {
			position178, tokenIndex178 := position, tokenIndex
			{
				position179 := position
				{
					position180, tokenIndex180 := position, tokenIndex
					if !matchDot() {
						goto l180
					}
					goto l178
				l180:
					position, tokenIndex = position180, tokenIndex180
				}
				add(ruleEOT, position179)
			}
			return true
		l178:
			position, tokenIndex = position178, tokenIndex178
			return false
		},
	}
	p.rules = _rules
	return nil
}