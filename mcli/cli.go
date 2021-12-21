package mcli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
)

// Command holds the information of a command.
type Command struct {
	Name        string
	Description string
	Hidden      bool

	f func()
}

type commands []*Command

func (p commands) sort() {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Name < p[j].Name
	})
}

func (p commands) search(cmdArgs []string) (cmd *Command, pars *parsing) {
	pars = &parsing{}
	flagIdx := len(cmdArgs)
	for i, x := range cmdArgs {
		if strings.HasPrefix(x, "-") {
			flagIdx = i
			break
		}
	}
	hasFlags := flagIdx != len(cmdArgs)
	if p[0].Name == "" {
		cmd = p[0]
	}
	cmdIdx := 0
	tryName := ""
	args := cmdArgs[:]
	for i := 0; i < len(cmdArgs); i++ {
		arg := cmdArgs[i]
		if strings.HasPrefix(arg, "-") {
			break
		}
		args = cmdArgs[i+1:]
		if tryName != "" {
			tryName += " "
		}
		tryName += arg
		idx := sort.Search(len(p), func(i int) bool {
			return p[i].Name >= tryName
		})
		if idx < len(p) && p[idx].Name == tryName {
			cmd = p[idx]
			pars.name = tryName
			cmdIdx = i
			continue
		}
		args = cmdArgs[i:]
		if hasFlags {
			invalidCmdName := strings.Join(cmdArgs[:flagIdx], " ")
			cmd = nil
			pars.name = invalidCmdName
			pars.args = &args
			return
		}
	}
	pars.args = &args
	argIdx := cmdIdx + 1
	if argIdx >= flagIdx {
		argIdx = flagIdx
	}
	pars.maybeArguments = cmdArgs[argIdx:flagIdx]
	pars.fs = flag.NewFlagSet("", flag.ExitOnError)
	return
}

func (p commands) listSubCommands(name string, filterHidden bool) (sub commands) {
	for _, cmd := range p {
		if cmd.Name != name && strings.HasPrefix(cmd.Name, name) {
			if cmd.Hidden && filterHidden {
				continue
			}
			sub = append(sub, cmd)
		}
	}
	return
}

var state struct {
	cmds commands
	*parsing
}

type parsing struct {
	name string
	args *[]string

	maybeArguments []string

	fs       *flag.FlagSet
	nonflags []*_flag
}

// Add adds a command to internal state.
func Add(name string, f func(), description string) {
	state.cmds = append(state.cmds, &Command{
		Name:        name,
		Description: description,
		f:           f,
	})
}

// AddHidden adds a command to internal state.
// A hidden command won't be showed in usage doc.
func AddHidden(name string, f func(), description string) {
	state.cmds = append(state.cmds, &Command{
		Name:        name,
		Description: description,
		Hidden:      true,
		f:           f,
	})
}

// AddGroup adds a group to internal state.
// A group is a common prefix for some commands.
func AddGroup(name string, description string) {
	state.cmds = append(state.cmds, &Command{
		Name:        name,
		Description: description,
		f:           groupCmd,
	})
}

var groupCmd = func() {
	emptyArgs := struct{}{}
	Parse(&emptyArgs)
	PrintHelp()
}

// Run runs the program, it will parse the command line and search
// for a registered command, it runs the command if a command is found,
// else it will report an error and exit the program.
func Run() {
	cmds := state.cmds
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name < cmds[j].Name
	})
	cmd, pars := cmds.search(os.Args[1:])
	if cmd == nil {
		if pars.name != "" {
			fmt.Fprintf(os.Stderr, "command not found: %s\n", pars.name)
		}
		printUsage(cmds, pars)()
		return
	}
	state.parsing = pars
	cmd.f()
}

func printAvailableCommands(out io.Writer, name string, cmds commands) {
	if sub := cmds.listSubCommands(name, true); len(sub) > 0 {
		cmds = sub
	}
	if len(cmds) == 0 {
		return
	}
	var cmdLines [][2]string
	prefix := []string{""}
	preName := ""
	for _, cmd := range cmds {
		if cmd.Name == "" || cmd.Hidden {
			continue
		}
		if preName != "" && cmd.Name != preName {
			if strings.HasPrefix(cmd.Name, preName) {
				prefix = append(prefix, preName)
			} else {
				for i := len(prefix) - 1; i > 0; i-- {
					if !strings.HasPrefix(cmd.Name, prefix[i]) {
						prefix = prefix[:i]
					}
				}
			}
		}
		leafCmdName := strings.TrimSpace(strings.TrimPrefix(cmd.Name, prefix[len(prefix)-1]))
		linePart0 := strings.Repeat("  ", len(prefix)) + leafCmdName
		linePart1 := cmd.Description
		cmdLines = append(cmdLines, [2]string{linePart0, linePart1})
		preName = cmd.Name
	}
	fmt.Fprint(out, "COMMANDS:\n")
	printWithPadding(out, cmdLines)
}

// Parse parses the command line for flags and arguments.
// v should be a pointer to a struct, else it panics.
func Parse(v interface{}, opts ...ParseOpt) (fs *flag.FlagSet, err error) {
	options := &parseOptions{
		errorHandling: flag.ExitOnError,
	}
	for _, o := range opts {
		o(options)
	}

	cmds := state.cmds
	pars := state.parsing
	if pars == nil {
		pars = &parsing{}
	}

	fs = flag.NewFlagSet("", options.errorHandling)
	nonflags := parseTags(fs, reflect.ValueOf(v).Elem())

	cmdName := options.cmdName
	if cmdName == "" {
		cmdName = pars.name
	}
	fs.Usage = printUsage(cmds, pars)
	pars.fs = fs
	pars.nonflags = nonflags

	if len(pars.maybeArguments) > 0 {
		if len(pars.nonflags) == 0 {
			invalidCmdName := cmdName
			if invalidCmdName != "" {
				invalidCmdName += " "
			}
			invalidCmdName += strings.Join(pars.maybeArguments, " ")
			failf(fs, &err, "command not found: %s", invalidCmdName)
			return
		}
	}

	osArgs := os.Args[1:]
	if pars.args != nil {
		osArgs = *pars.args
	}
	if options.args != nil {
		osArgs = *options.args
	}

	if err = fs.Parse(osArgs); err != nil {
		return fs, err
	}
	if err = parseNonflags(fs, pars.nonflags, fs.Args()); err != nil {
		return fs, err
	}
	if err = checkRequired(fs, pars.nonflags); err != nil {
		return fs, err
	}
	return fs, err
}

func parseNonflags(fs *flag.FlagSet, nonflags []*_flag, args []string) (err error) {
	i, j := 0, 0
	for i < len(nonflags) && j < len(args) {
		f := nonflags[i]
		value := args[j]
		e := f.Set(value)
		if e != nil {
			failf(fs, &err, "invalid value %q for argument %s: %v", value, f.name, e)
			break
		}
		if f.rv.Kind() != reflect.Slice {
			i++
		}
		j++
	}
	return
}

func checkRequired(fs *flag.FlagSet, nonflags []*_flag) (err error) {
	fs.VisitAll(func(ff *flag.Flag) {
		if err == nil {
			f := ff.Value.(*_flag)
			if f.required && reflect.Indirect(f.rv).IsZero() {
				failf(fs, &err, "required flag not set: %v", f.name)
			}
		}
	})
	if err == nil {
		for _, f := range nonflags {
			if f.required && reflect.Indirect(f.rv).IsZero() {
				failf(fs, &err, "required argument not set: %v", f.name)
				break
			}
		}
	}
	return
}

func failf(fs *flag.FlagSet, errp *error, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(fs.Output(), msg)
	fs.Usage()
	switch fs.ErrorHandling() {
	case flag.ExitOnError:
		os.Exit(2)
	case flag.ContinueOnError:
		if *errp == nil {
			*errp = errors.New(msg)
		}
	case flag.PanicOnError:
		panic(msg)
	}
}

type parseOptions struct {
	cmdName       string
	args          *[]string
	errorHandling flag.ErrorHandling
}

// ParseOpt specifies options to change the behavior of Parse.
type ParseOpt func(*parseOptions)

// WithArgs indicates Parse to parse from the given args, instead of
// parsing from the program's command line arguments.
func WithArgs(args []string) ParseOpt {
	return func(options *parseOptions) {
		options.args = &args
	}
}

// WithErrorHandling indicates Parse to use the given ErrorHandling.
// By default, Parse exits the program when an error happens.
func WithErrorHandling(h flag.ErrorHandling) ParseOpt {
	return func(options *parseOptions) {
		options.errorHandling = h
	}
}

// WithName specifies the name to use when printing usage doc.
func WithName(name string) ParseOpt {
	return func(options *parseOptions) {
		options.cmdName = name
	}
}

func parseTags(fs *flag.FlagSet, rv reflect.Value) (nonflags []*_flag) {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Field(i)
		ft := rt.Field(i)
		cliTag := ft.Tag.Get("cli")
		defaultValue := ft.Tag.Get("default")
		if isIgnoreTag(cliTag) {
			continue
		}
		if fv.Kind() == reflect.Struct {
			parseTags(fs, fv)
			continue
		}
		if cliTag == "" {
			continue
		}
		f := parseFlag(cliTag, defaultValue, fv)
		if f == nil || f.name == "" {
			continue
		}
		if f.nonflag {
			nonflags = append(nonflags, f)
			continue
		}
		fs.Var(f, f.name, f.description)
		if f.short != "" {
			fs.Var(f, f.short, f.description)
		}
	}
	return
}

func printUsage(cmds commands, p *parsing) func() {
	return func() {
		fs := p.fs
		cmdName := p.name
		nonflags := p.nonflags

		flagCount := 0
		hasShortFlag := false
		fs.VisitAll(func(ff *flag.Flag) {
			f := ff.Value.(*_flag)
			if !f.hidden {
				flagCount++
				hasShortFlag = hasShortFlag || f.short != ""
			}
		})
		subCmds := cmds.listSubCommands(cmdName, true)

		out := fs.Output()
		progName := path.Base(os.Args[0])
		hasFlags, hasNonflags := flagCount > 0, len(nonflags) > 0
		hasSubCmds := len(subCmds) > 0
		usage := "USAGE:\n  " + progName
		if cmdName != "" {
			usage += " " + cmdName
		}
		if hasFlags {
			usage += " [flags]"
		}
		if hasNonflags {
			for _, f := range nonflags {
				name := f.name
				if f.isSlice() {
					name += "..."
				}
				if f.required {
					usage += fmt.Sprintf(" <%s>", name)
				} else {
					usage += fmt.Sprintf(" [%s]", name)
				}
			}
		}
		if !hasFlags && !hasNonflags && hasSubCmds {
			usage += " <command> ..."
		}
		fmt.Fprint(out, usage, "\n\n")

		if flagCount > 0 {
			var flagLines [][2]string
			fs.VisitAll(func(ff *flag.Flag) {
				f := ff.Value.(*_flag)
				if f.hidden {
					return
				}
				if f.name != ff.Name {
					return
				}
				prefix, usage := f.getUsage(hasShortFlag)
				flagLines = append(flagLines, [2]string{prefix, usage})
			})
			fmt.Fprint(out, "FLAGS:\n")
			printWithPadding(out, flagLines)
			fmt.Fprint(out, "\n")
		}

		if len(nonflags) > 0 {
			var nonflagLines [][2]string
			for _, f := range nonflags {
				prefix, usage := f.getUsage(false)
				nonflagLines = append(nonflagLines, [2]string{prefix, usage})
			}
			fmt.Fprint(out, "ARGUMENTS:\n")
			printWithPadding(out, nonflagLines)
			fmt.Fprint(out, "\n")
		}

		if hasSubCmds {
			printAvailableCommands(out, cmdName, cmds)
			fmt.Fprint(out, "\n")
		}
	}
}

func printWithPadding(out io.Writer, lines [][2]string) {
	const _N = 30
	maxPrefixLen := 0
	for _, line := range lines {
		if n := len(line[0]); n > maxPrefixLen && n <= _N {
			maxPrefixLen = n
		}
	}
	for _, line := range lines {
		x, y := line[0], line[1]
		fmt.Fprint(out, x)
		if y != "" {
			if len(x) < _N {
				fmt.Fprint(out, strings.Repeat(" ", maxPrefixLen+4-len(x)))
				fmt.Fprint(out, strings.ReplaceAll(y, "\n", "\n    \t"))
			} else {
				fmt.Fprint(out, "\n    \t")
				fmt.Fprint(out, strings.ReplaceAll(y, "\n", "\n    \t"))
			}
		}
		fmt.Fprint(out, "\n")
	}
}

// PrintHelp prints usage doc of the current command to stderr.
func PrintHelp() {
	cmds := state.cmds
	pars := state.parsing
	if pars == nil {
		pars = &parsing{fs: &flag.FlagSet{}}
	}
	printUsage(cmds, pars)()
}
