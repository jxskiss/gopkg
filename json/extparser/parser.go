package extparser

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"unsafe"
)

//go:generate peg json.peg

const maxImportDepth = 10

func Parse(data []byte, importRoot string) ([]byte, error) {
	return parse(data, importRoot, 0)
}

func parse(data []byte, importRoot string, depth int) ([]byte, error) {
	if depth > maxImportDepth {
		return nil, errors.New("max import depth exceeded")
	}

	doc := &JSON{
		Buffer: b2s(data),
	}
	if err := doc.Init(); err != nil {
		return nil, err
	}
	if err := doc.Parse(); err != nil {
		return nil, err
	}
	if !doc.hasExtendedFeature() {
		return data, nil
	}

	parser := &parser{
		doc:   doc,
		buf:   make([]byte, 0, len(data)),
		root:  importRoot,
		depth: depth,
	}
	return parser.rewrite()
}

type parser struct {
	doc *JSON
	buf []byte

	root  string
	depth int
}

func (p *parser) text(n *node32) string {
	return string(p.doc.buffer[n.begin:n.end])
}

func (p *parser) rewrite() ([]byte, error) {
	root := p.doc.AST()
	if root.pegRule != ruleDocument {
		return nil, errors.New("invalid JSON document")
	}

	for n := root.up; n != nil; n = n.next {
		switch n.pegRule {
		case ruleSpacing:
			continue
		case ruleJSON:
			if err := p.parseJSON(n); err != nil {
				return nil, err
			}
		}
	}
	return p.buf, nil
}

func (p *parser) parseJSON(n *node32) (err error) {
	n = n.up
	switch n.pegRule {
	case ruleObject:
		if err = p.parseObject(n); err != nil {
			return
		}
	case ruleArray:
		if err = p.parseArray(n); err != nil {
			return
		}
	case ruleString:
		p.buf = append(p.buf, p.parseString(n)...)
	case ruleTrue:
		p.buf = append(p.buf, "true"...)
	case ruleFalse:
		p.buf = append(p.buf, "false"...)
	case ruleNull:
		p.buf = append(p.buf, "null"...)
	case ruleNumber:
		p.buf = append(p.buf, p.text(n)...)
	case ruleImport:
		if err = p.parseImport(n); err != nil {
			return
		}
	}
	return nil
}

func (p *parser) parseObject(n *node32) (err error) {
	var preRule pegRule
	for n := n.up; n != nil; n = n.next {
		switch n.pegRule {
		case ruleLWING:
			p.buf = append(p.buf, '{')
		case ruleRWING:
			if preRule == ruleCOMMA {
				p.buf = p.buf[:len(p.buf)-1]
			}
			p.buf = append(p.buf, '}')
		case ruleCOLON:
			p.buf = append(p.buf, ':')
		case ruleCOMMA:
			p.buf = append(p.buf, ',')
		case ruleString:
			p.buf = append(p.buf, p.parseString(n)...)
		case ruleJSON:
			err = p.parseJSON(n)
			if err != nil {
				return
			}
		}
		preRule = n.pegRule
	}
	return nil
}

func (p *parser) parseArray(n *node32) (err error) {
	var preRule pegRule
	for n := n.up; n != nil; n = n.next {
		switch n.pegRule {
		case ruleLBRK:
			p.buf = append(p.buf, '[')
		case ruleRBRK:
			if preRule == ruleCOMMA {
				p.buf = p.buf[:len(p.buf)-1]
			}
			p.buf = append(p.buf, ']')
		case ruleCOMMA:
			p.buf = append(p.buf, ',')
		case ruleJSON:
			err = p.parseJSON(n)
			if err != nil {
				return
			}
		}
		preRule = n.pegRule
	}
	return nil
}

func (p *parser) parseString(n *node32) string {
	n = n.up
	switch n.pegRule {
	case ruleSingleQuoteLiteral:
		return "\"" + string(p.doc.buffer[n.begin+1:n.end-1]) + "\""
	case ruleDoubleQuoteLiteral:
		return p.text(n)
	}
	return ""
}

func (p *parser) parseImport(n *node32) (err error) {
	n = n.up
	importPath := p.parseString(n)
	importPath = filepath.Join(p.root, importPath[1:len(importPath)-1])
	included, err := ioutil.ReadFile(importPath)
	if err != nil {
		return
	}
	included, err = parse(included, p.root, p.depth+1)
	if err != nil {
		return
	}
	p.buf = append(p.buf, included...)
	return nil
}

func (p *JSON) hasExtendedFeature() bool {
	var preRule pegRule
	for _, n := range p.Tokens() {
		switch n.pegRule {
		case ruleSingleQuoteLiteral,
			ruleImport,
			ruleLongComment, ruleLineComment, rulePragma:
			return true
		case ruleRWING:
			if preRule == ruleCOMMA {
				return true
			}
		case ruleRBRK:
			if preRule == ruleCOMMA {
				return true
			}
		case ruleTrue:
			if string(p.buffer[n.begin:n.end]) != "true" {
				return true
			}
		case ruleFalse:
			if string(p.buffer[n.begin:n.end]) != "false" {
				return true
			}
		case ruleNull:
			if string(p.buffer[n.begin:n.end]) != "null" {
				return true
			}
		}
		preRule = n.pegRule
	}
	return false
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
