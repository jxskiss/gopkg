package yamlx

import (
	"errors"
	"fmt"
	"strings"
)

const (
	directiveEnv      = "@@env"
	directiveVariable = "@@var"
	directiveInclude  = "@@incl"
	directiveRefer    = "@@ref"
	directiveFunction = "@@fn"
)

type directive struct {
	name string
	args map[string]any
}

func (d *directive) getRefPath(nodePfx []string) (path, origPath string, isTostr bool, modifier string) {
	origPath = d.args["path"].(string)
	isRelative, level, tail := isRelativeJSONPath(origPath)
	path = tail
	if isRelative {
		prefix := strings.Join(nodePfx[:max(len(nodePfx)-level, 0)], ".")
		if prefix != "" {
			path = prefix + "." + tail
		}
	}
	isTostr, modifier = hasTostrModifier(path)
	if isTostr {
		path = strings.TrimSuffix(path, modifier)
	}
	return path, origPath, isTostr, modifier
}

func parseDirective(str string) (d directive, ok bool, err error) {
	if !strings.HasPrefix(str, "@@") {
		return
	}
	switch {
	case hasDirectivePrefix(str, directiveEnv):
		d, err = parseEnvDirective(str)
	case hasDirectivePrefix(str, directiveVariable):
		d, err = parseVariableDirective(str)
	case hasDirectivePrefix(str, directiveInclude):
		d, err = parseIncludeDirective(str)
	case hasDirectivePrefix(str, directiveRefer):
		d, err = parseReferDirective(str)
	case hasDirectivePrefix(str, directiveFunction):
		d, err = parseFunctionDirective(str)
	default:
		err = fmt.Errorf("unrecognized directive: %q", str)
		return directive{}, false, err
	}
	ok = err == nil
	return
}

func hasDirectivePrefix(str, directive string) bool {
	return strings.HasPrefix(str, directive+" ")
}

func parseEnvDirective(str string) (directive, error) {
	str = strings.TrimPrefix(str, directiveEnv)
	str = strings.TrimSpace(str)

	var envNames []string
	for _, x := range strings.Split(str, ",") {
		if x = strings.TrimSpace(x); x != "" {
			envNames = append(envNames, x)
		}
	}
	if len(envNames) == 0 {
		err := errors.New("missing environment variable name for @@env directive")
		return directive{}, err
	}
	args := map[string]any{
		"envNames": envNames,
	}
	return directive{name: directiveEnv, args: args}, nil
}

func parseVariableDirective(str string) (directive, error) {
	str = strings.TrimPrefix(str, directiveVariable)
	str = strings.TrimSpace(str)
	if str == "" {
		err := errors.New("missing variable name for @@var directive")
		return directive{}, err
	}
	args := map[string]any{
		"varName": str,
	}
	return directive{name: directiveVariable, args: args}, nil
}

func parseIncludeDirective(str string) (directive, error) {
	str = strings.TrimPrefix(str, directiveInclude)
	filename := strings.TrimSpace(str)
	if filename == "" {
		err := errors.New("missing filename for @@inc directive")
		return directive{}, err
	}
	args := map[string]any{
		"filename": filename,
	}
	return directive{name: directiveInclude, args: args}, nil
}

func parseReferDirective(str string) (directive, error) {
	str = strings.TrimPrefix(str, directiveRefer)
	str = strings.TrimSpace(str)
	if str == "" {
		err := errors.New("missing JSON path for @@ref directive")
		return directive{}, err
	}
	args := map[string]any{
		"path": str,
	}
	return directive{name: directiveRefer, args: args}, nil
}

func isRelativeJSONPath(path string) (ok bool, level int, tail string) {
	if path[0] == '.' &&
		strings.HasPrefix(strings.TrimLeft(path, "."), "/") {
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 {
			return true, len(parts[0]), parts[1]
		}
	}
	return false, 0, path
}

func hasTostrModifier(path string) (ok bool, modifier string) {
	pos := strings.Index(path, "@tostr")
	if pos > 1 {
		if path[pos-1] == '|' || path[pos-1] == '.' {
			return true, path[pos-1:]
		}
	}
	return false, ""
}

func parseFunctionDirective(str string) (directive, error) {
	str = strings.TrimPrefix(str, directiveFunction)
	str = strings.TrimSpace(str)
	if str == "" {
		err := errors.New("missing function expression for @@fn directive")
		return directive{}, err
	}
	args := map[string]any{
		"expr": str,
	}
	return directive{name: directiveFunction, args: args}, nil
}

//nolint:unused
func trimParensAndSpace(str string) string {
	if str != "" && str[0] == '(' && str[len(str)-1] == ')' {
		str = str[1 : len(str)-1]
		str = strings.TrimSpace(str)
	}
	return str
}

//nolint:unused
func trimQuotAndSpace(str string) string {
	if str != "" {
		if (str[0] == '"' && str[len(str)-1] == '"') ||
			(str[0] == '\'' && str[len(str)-1] == '\'') {
			str = str[1 : len(str)-1]
			str = strings.TrimSpace(str)
		}
	}
	return str
}

func unescapeStrValue(str string) string {
	bsCount := 0
	isAtAt := false
	if strings.HasPrefix(str, "\\") {
		for i := 0; i < len(str); i++ {
			if str[i] == '\\' {
				bsCount++
				continue
			}
			isAtAt = strings.HasPrefix(str[i:], "@@")
			break
		}
	}
	if bsCount == 0 || !isAtAt {
		return str
	}
	return str[:bsCount-1] + str[bsCount:]
}
