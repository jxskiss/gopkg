package yamlx

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/json"
	"github.com/jxskiss/gopkg/v2/utils/strutil"
)

const strTag = "!!str"

type nodeStack[T any] []T

func (p *nodeStack[T]) push(nodes ...T) {
	*p = append(*p, nodes...)
}

func (p *nodeStack[T]) pop() (top T) {
	if len(*p) == 0 {
		return top
	}
	top = (*p)[len(*p)-1]
	*p = (*p)[:len(*p)-1]
	return
}

type pathTuple struct{ path, origPath string }

func (p pathTuple) String() string {
	if p.path == p.origPath {
		return p.path
	}
	return fmt.Sprintf("%s (%s)", p.origPath, p.path)
}

type parser struct {
	parsed   bool
	parseErr error

	filename string
	opts     *extOptions
	buf      []byte
	doc      *yaml.Node

	// directive inc
	incStack []string

	// directive ref
	refMark     string
	refCounter  int
	refTable    map[string]int
	refRevTable map[int]pathTuple
	refDag      dag

	// directive var
	varMap     map[string]*yaml.Node
	varNodeMap map[*yaml.Node]string

	// directive fn
	funcValMap map[string]reflect.Value
}

func newParser(buf []byte, options ...Option) *parser {
	opts := new(extOptions).apply(options...)
	p := &parser{opts: opts, buf: buf}
	p.addFuncs(opts.FuncMap)
	return p
}

func newParserWithOpts(buf []byte, opts *extOptions) *parser {
	p := &parser{opts: opts, buf: buf}
	p.addFuncs(opts.FuncMap)
	return p
}

func (p *parser) Unmarshal(v any) error {
	err := p.parse()
	if err != nil {
		return err
	}

	// Unescape string values before unmarshalling.
	p.unescapeStrings()

	return p.doc.Decode(v)
}

func (p *parser) parse() (err error) {
	if p.parsed {
		return p.parseErr
	}
	defer func() {
		p.parseErr = err
		p.parsed = true
	}()

	if len(p.buf) == 0 && p.filename != "" {
		buf, err := os.ReadFile(p.filename)
		if err != nil {
			return fmt.Errorf("cannot read file: %w", err)
		}
		p.buf = buf
	}

	p.doc = &yaml.Node{}
	err = yaml.Unmarshal(p.buf, p.doc)
	if err != nil {
		return err
	}

	// The env, fn and var directives are scoped within a single file,
	// they should be resolved before the include directives.
	if err = p.resolveEnvAndFunctions(); err != nil {
		return err
	}
	if err = p.resolveVariables(); err != nil {
		return err
	}

	// resolve includes
	if p.opts.EnableInclude {
		if err = p.resolveIncludes(); err != nil {
			return err
		}
	}

	// The ref directives are allowed to reference data from included
	// files, they should be resolved after the include directives.
	if err = p.resolveReferences(); err != nil {
		return err
	}

	return nil
}

func (p *parser) resolveEnvAndFunctions() error {
	if p.doc == nil {
		return nil
	}

	// depth-first traversal
	stack := make(nodeStack[*yaml.Node], 0, 64)
	stack.push(p.doc)
	for len(stack) > 0 {
		node := stack.pop()
		if node == nil || node.IsZero() {
			continue
		}
		switch node.Kind {
		case yaml.DocumentNode:
			if len(node.Content) == 0 || node.Content[0].IsZero() {
				continue
			}
			stack.push(node.Content[0])
		case yaml.SequenceNode:
			stack.push(node.Content...)
		case yaml.MappingNode:
			for i, j := 0, 1; i < len(node.Content); i, j = i+2, j+2 {
				stack.push(node.Content[j])
			}
		case yaml.AliasNode:
			continue
		case yaml.ScalarNode:
			if node.Tag != strTag {
				continue
			}
			directive, ok, err := parseDirective(node.Value)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			if p.opts.EnableEnv && directive.name == directiveEnv {
				envNames := directive.args["envNames"].([]string)
				readEnv(node, envNames)
				continue
			}
			if directive.name == directiveFunction {
				fnRet, err := p.callFunction(directive.args["expr"].(string))
				if err != nil {
					return err
				}
				newNode, err := convToNode(fnRet)
				if err != nil {
					return err
				}
				newNode.LineComment = node.LineComment
				*node = *newNode
				continue
			}
		}
	}
	return nil
}

func readEnv(node *yaml.Node, envNames []string) {
	found := false
	for _, name := range envNames {
		val := os.Getenv(name)
		if val != "" {
			found = true
			node.Value = val
			break
		}
	}
	if !found {
		node.Value = ""
	}
}

func convToNode(value any) (*yaml.Node, error) {
	// Use marshal and unmarshal to avoid string escaping issues.
	var node = &yaml.Node{}
	buf, err := yaml.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal function result: %w", err)
	}
	err = yaml.Unmarshal(buf, node)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal function result to node: %w", err)
	}
	if node.Kind == yaml.DocumentNode {
		node = node.Content[0]
	}
	return node, nil
}

func (p *parser) resolveVariables() error {
	if p.doc == nil {
		return nil
	}

	p.varMap = make(map[string]*yaml.Node)
	p.varNodeMap = make(map[*yaml.Node]string)

	// depth-first traversal
	stack := make(nodeStack[*yaml.Node], 0, 64)
	stack.push(p.doc)
	for len(stack) > 0 {
		node := stack.pop()
		if node == nil || node.IsZero() {
			continue
		}
		switch node.Kind {
		case yaml.DocumentNode:
			if len(node.Content) == 0 || node.Content[0].IsZero() {
				continue
			}
			stack.push(node.Content[0])
		case yaml.SequenceNode:
			for i, n := range node.Content {
				// 当 array 有 anchor 时，lineComment 会被算给第一个列表元素，
				// 如果第一个列表元素也有 lineComment，则无法正确区分到底是哪一样的注释，
				// 因此不支持带有 anchor 的 array 作为变量目标对象。
				if i == 0 && node.Anchor != "" && n.LineComment != "" {
					if directive, _ := parseVariableDirective(n.LineComment); directive.name == directiveVariable {
						varName := directive.args["varName"].(string)
						return fmt.Errorf("mix using anchor and @@var directive does not work correctly and is not supported: %v", varName)
					}
				}
			}

			stack.push(node.Content...)
			if err := p.checkAndAddVariable(node.LineComment, node); err != nil {
				return err
			}
		case yaml.MappingNode:
			if err := p.checkAndAddVariable(node.LineComment, node); err != nil {
				return err
			}
			for i, j := 0, 1; j < len(node.Content); i, j = i+2, j+2 {
				kNode := node.Content[i]
				vNode := node.Content[j]

				// 当 map 有 anchor 时，lineComment 会被算给第一个 kv，
				// 然而若第一个 kv 是分行书写的且 key 后面又跟了注释，则无法正确区分到底是哪一行的注释，
				// 因此不支持带有 anchor 的 map 作为变量目标对象。
				if i == 0 && node.Anchor != "" && kNode.LineComment != "" {
					if directive, _ := parseVariableDirective(kNode.LineComment); directive.name == directiveVariable {
						varName := directive.args["varName"].(string)
						return fmt.Errorf("mix using anchor and @@var directive does not work correctly and is not supported: %v", varName)
					}
				}

				stack.push(vNode)
				if err := p.checkAndAddVariable(kNode.LineComment, vNode); err != nil {
					return err
				}
			}
		case yaml.AliasNode:
			if err := p.checkAndAddVariable(node.LineComment, node); err != nil {
				return err
			}
			continue
		case yaml.ScalarNode:
			if err := p.checkAndAddVariable(node.LineComment, node); err != nil {
				return err
			}
			if node.Tag == strTag {
				directive, ok, err := parseDirective(node.Value)
				if err != nil {
					return err
				}
				if ok && directive.name == directiveVariable {
					p.varNodeMap[node] = directive.args["varName"].(string)
				}
			}
		}
	}
	if len(p.varNodeMap) == 0 {
		return nil
	}

	for node, varName := range p.varNodeMap {
		dstNode := p.varMap[varName]
		if dstNode == nil {
			return fmt.Errorf("undefined variable: %s", varName)
		}
		*node = *dstNode
	}

	cyclicVarName, isCyclic := p.detectVarCircle(p.doc, nil)
	if isCyclic {
		return fmt.Errorf("circular variable reference detected: %s", cyclicVarName)
	}
	return nil
}

func (p *parser) checkAndAddVariable(lineComment string, node *yaml.Node) error {
	lineComment = strings.TrimLeft(lineComment, "#")
	lineComment = strings.TrimSpace(lineComment)
	if !strings.HasPrefix(lineComment, directiveVariable) {
		return nil
	}
	directive, err := parseVariableDirective(lineComment)
	if err != nil {
		return err
	}
	varName := directive.args["varName"].(string)
	p.varMap[varName] = node
	return nil
}

func (p *parser) detectVarCircle(node *yaml.Node, stack nodeStack[*yaml.Node]) (string, bool) {
	if node == nil || node.IsZero() {
		return "", false
	}

	for _, n := range stack {
		if n == node {
			var varName string
			revVarMap := make(map[*yaml.Node]string)
			for name, varNode := range p.varMap {
				revVarMap[varNode] = name
			}
			for _, seenNode := range stack {
				if name := p.varNodeMap[seenNode]; name != "" {
					varName = name
				} else if name = revVarMap[seenNode]; name != "" {
					varName = name
				}
			}
			return varName, true
		}
	}

	stack.push(node)
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 || node.Content[0].IsZero() {
			return "", false
		}
		return p.detectVarCircle(node.Content[0], stack)
	case yaml.SequenceNode:
		for _, elemNode := range node.Content {
			if varName, isCyclic := p.detectVarCircle(elemNode, stack); isCyclic {
				return varName, true
			}
		}
	case yaml.MappingNode:
		for i, j := 0, 1; j < len(node.Content); i, j = i+2, j+2 {
			elemNode := node.Content[j]
			if varName, isCyclic := p.detectVarCircle(elemNode, stack); isCyclic {
				return varName, true
			}
		}
	case yaml.AliasNode, yaml.ScalarNode:
		return "", false
	}
	return "", false
}

func (p *parser) resolveIncludes() error {
	if p.doc == nil {
		return nil
	}

	// depth-first traversal
	stack := make(nodeStack[*yaml.Node], 0, 64)
	stack.push(p.doc)
	for len(stack) > 0 {
		node := stack.pop()
		if node == nil || node.IsZero() {
			continue
		}
		switch node.Kind {
		case yaml.DocumentNode:
			if len(node.Content) == 0 || node.Content[0].IsZero() {
				continue
			}
			stack.push(node.Content[0])
		case yaml.SequenceNode:
			stack.push(node.Content...)
		case yaml.MappingNode:
			for i, j := 0, 1; j < len(node.Content); i, j = i+2, j+2 {
				stack.push(node.Content[j])
			}
		case yaml.AliasNode:
			continue
		case yaml.ScalarNode:
			if node.Tag != strTag {
				continue
			}
			directive, ok, err := parseDirective(node.Value)
			if err != nil {
				return err
			}
			if !ok || directive.name != directiveInclude {
				continue
			}

			// Execute the include directive.
			incFilePath, err := p.getIncludeAbsFilename(directive.args["filename"].(string))
			if err != nil {
				return err
			}
			for _, fn := range p.incStack {
				if fn == incFilePath {
					return fmt.Errorf("circular include detected: %s", incFilePath)
				}
			}
			incBuf, err := os.ReadFile(incFilePath)
			if err != nil {
				return fmt.Errorf("cannot read include file: %w", err)
			}
			incParser := newParserWithOpts(incBuf, p.opts)
			incParser.filename = incFilePath
			incParser.incStack = append(clip(p.incStack), incFilePath)
			err = incParser.parse()
			if err != nil {
				return fmt.Errorf("cannot parse include file: %w", err)
			}
			*node = *(incParser.getDocValueNode())
		}
	}

	return nil
}

func (p *parser) getIncludeAbsFilename(s string) (string, error) {
	if filepath.IsAbs(s) {
		return filepath.Clean(s), nil
	}

	var includeDirs []string
	if p.filename != "" {
		isRelative := strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../")
		dir := filepath.Dir(p.filename)
		if isRelative {
			includeDirs = append([]string{dir}, p.opts.IncludeDirs...)
		} else {
			includeDirs = append(p.opts.IncludeDirs, dir)
		}
	} else {
		includeDirs = p.opts.IncludeDirs
	}

	var filename string
	for _, dir := range includeDirs {
		fName := filepath.Join(dir, s)
		info, err := os.Stat(fName)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("error checking include file: %w", err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("include file is a directory: %v", fName)
		}
		filename = fName
		break
	}
	if filename != "" {
		return filename, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get working directory: %w", err)
	}
	return filepath.Abs(filepath.Join(wd, s))
}

func (p *parser) resolveReferences() error {
	if p.doc == nil {
		return nil
	}

	toStrRefs := make(map[int]string) // seq -> modifier

	// depth-first traversal
	type NodePath struct {
		N *yaml.Node // node
		P []string   // path prefix
	}
	stack := make(nodeStack[NodePath], 0, 64)
	stack.push(NodePath{p.doc, nil})
	for len(stack) > 0 {
		node := stack.pop()
		if node.N == nil || node.N.IsZero() {
			continue
		}
		switch node.N.Kind {
		case yaml.DocumentNode:
			if len(node.N.Content) == 0 || node.N.Content[0].IsZero() {
				continue
			}
			stack.push(NodePath{node.N.Content[0], nil})
		case yaml.SequenceNode:
			for i := 0; i < len(node.N.Content); i++ {
				_n := node.N.Content[i]
				_p := append(clip(node.P), strconv.Itoa(i))
				stack.push(NodePath{_n, _p})
			}
		case yaml.MappingNode:
			for i, j := 0, 1; j < len(node.N.Content); i, j = i+2, j+2 {
				_n := node.N.Content[j]
				_p := append(clip(node.P), gjson.Escape(node.N.Content[i].Value))
				stack.push(NodePath{_n, _p})
			}
		case yaml.AliasNode:
			continue
		case yaml.ScalarNode:
			if node.N.Tag != strTag {
				continue
			}
			directive, ok, err := parseDirective(node.N.Value)
			if err != nil {
				return err
			}
			if !ok || directive.name != directiveRefer {
				continue
			}

			// Note we need special processing for modifier "@tostr".
			jsonPath, origPath, isTostr, modifier := directive.getRefPath(node.P)
			seq, placeholder := p.getReferID(jsonPath, origPath)
			node.N.Value = placeholder
			p.refDag.addVertex(seq)
			if isTostr {
				toStrRefs[seq] = modifier
			}
		}
	}
	if p.refMark == "" {
		return nil
	}

	var intermediateValue any
	err := p.doc.Decode(&intermediateValue)
	if err != nil {
		return fmt.Errorf("cannot decode intermediate data: %w", err)
	}
	intermediateBuf, err := json.Marshal(intermediateValue)
	if err != nil {
		return fmt.Errorf("cannot marshal intermediat data: %w", err)
	}

	// Resolve dependency and do replacing.
	mark := p.refMark
	resolved := make(map[int]string, len(p.refTable))
	for seq := 1; seq <= len(p.refTable); seq++ {
		refPath := p.refRevTable[seq]
		r := gjson.GetBytes(intermediateBuf, refPath.path)
		if !r.Exists() {
			return fmt.Errorf("cannot find referenced data: %v", refPath)
		}
		resolved[seq] = r.Raw
		pos := 0
		for pos < len(r.Raw) {
			raw := r.Raw[pos:]
			idx := strings.Index(raw, mark)
			if idx < 0 {
				break
			}
			end := idx + len(mark) + 2
			for end < len(raw) {
				if raw[end] >= '0' && raw[end] <= '9' {
					end++
					continue
				}
				break
			}
			refSeqStr := raw[idx+len(mark)+1 : end]
			refSeq, err := strconv.Atoi(refSeqStr)
			if err != nil {
				return fmt.Errorf("invalid refer id: %w", err)
			}
			if refSeq == seq {
				return fmt.Errorf("circular reference detected: %s", refPath)
			}
			isCyclic := p.refDag.addEdge(refSeq, seq)
			if isCyclic {
				return fmt.Errorf("circular reference detected: %s", refPath)
			}
			pos += end
		}
	}

	order := p.refDag.topoSort()
	for _, seq := range order {
		if modifier := toStrRefs[seq]; modifier != "" {
			resolved[seq] = convToStr(resolved[seq], modifier)
		}
		final := resolved[seq]
		placeholder := `"` + p.referPlaceholder(seq) + `"`
		p.refDag.visitNeighbors(seq, func(to int) {
			resolved[to] = strings.Replace(resolved[to], placeholder, final, -1)
		})
	}
	oldnew := make([]string, 0, 2*len(resolved))
	for seq, text := range resolved {
		placeholder := `"` + p.referPlaceholder(seq) + `"`
		oldnew = append(oldnew, placeholder, text)
	}
	replacer := strings.NewReplacer(oldnew...)
	finalBuf := replacer.Replace(unsafeheader.BytesToString(intermediateBuf))
	p.doc = &yaml.Node{}
	return yaml.Unmarshal(unsafeheader.StringToBytes(finalBuf), p.doc)
}

func convToStr(value, modifier string) string {
	tmp := fmt.Sprintf(`{"a":%s}`, value)
	return gjson.Get(tmp, "a"+modifier).Raw
}

func (p *parser) getReferID(path, origPath string) (int, string) {
	if p.refMark == "" {
		p.refMark = strutil.RandomHex(40)
		p.refTable = make(map[string]int)
		p.refRevTable = make(map[int]pathTuple)
	}
	seq := p.refTable[path]
	if seq == 0 {
		p.refCounter++
		seq = p.refCounter
		p.refTable[path] = seq
		p.refRevTable[seq] = pathTuple{path, origPath}
	}
	placeholder := p.referPlaceholder(seq)
	return seq, placeholder
}

func (p *parser) referPlaceholder(n int) string {
	return fmt.Sprintf("%s_%d", p.refMark, n)
}

func (p *parser) getDocValueNode() *yaml.Node {
	if p.doc == nil {
		return &yaml.Node{}
	}
	switch p.doc.Kind {
	case yaml.DocumentNode:
		if len(p.doc.Content) > 0 {
			return p.doc.Content[0]
		}
		return nil
	default:
		return p.doc
	}
}

func (p *parser) unescapeStrings() {
	if p.doc == nil {
		return
	}

	// depth-first traversal
	stack := make(nodeStack[*yaml.Node], 0, 64)
	stack.push(p.doc)
	for len(stack) > 0 {
		node := stack.pop()
		if node == nil || node.IsZero() {
			continue
		}
		switch node.Kind {
		case yaml.DocumentNode:
			if len(node.Content) == 0 || node.Content[0].IsZero() {
				continue
			}
			stack.push(node.Content[0])
		case yaml.SequenceNode:
			stack.push(node.Content...)
		case yaml.MappingNode:
			stack.push(node.Content...)
		case yaml.AliasNode:
			continue
		case yaml.ScalarNode:
			if node.Tag != strTag {
				continue
			}
			node.Value = unescapeStrValue(node.Value)
		}
	}
}

func clip[T any](s []T) []T {
	return s[:len(s):len(s)]
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
