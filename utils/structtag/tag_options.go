package structtag

import "strings"

// Options represents a set of parsed options of a struct field tag.
type Options []Option

func (p Options) Get(option string) (Option, bool) {
	for _, opt := range p {
		if opt.k == option {
			return opt, true
		}
	}
	return Option{}, false
}

// Option represents a single option from a struct field tag.
type Option struct {
	raw, k, v string
}

// String returns the original string represent of the option.
func (p Option) String() string { return p.raw }

// Key returns the parsed key of the option, if available.
func (p Option) Key() string { return p.k }

// Value returns the parsed value of the option, if available.
func (p Option) Value() string { return p.v }

// ParseOptions parses tag into Options using optionSep and kvSep.
//
// If optionSep is not empty, it splits tag into options using optionSep
// as separator, else the whole tag is considered as a single option.
// If kvSep is not empty, it splits each option into key value pair using
// kvSep as separator, else the option's key, value will be empty.
func ParseOptions(tag string, optionSep, kvSep string) Options {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil
	}

	var options = make([]Option, 0, 4)
	var opts []string
	if optionSep == "" {
		opts = []string{tag}
	} else {
		opts = strings.Split(tag, optionSep)
	}

	for _, optstr := range opts {
		optstr = strings.TrimSpace(optstr)
		opt := Option{raw: optstr}
		if kvSep != "" {
			sepIdx := strings.Index(optstr, kvSep)
			if sepIdx < 0 {
				opt.k = optstr
			} else {
				opt.k = strings.TrimSpace(optstr[:sepIdx])
				opt.v = strings.TrimSpace(optstr[sepIdx+1:])
			}
		}
		options = append(options, opt)
	}
	return options
}
