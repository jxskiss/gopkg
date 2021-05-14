package structtag

import "strings"

type Options []Option

func (p Options) Get(option string) (Option, bool) {
	for _, opt := range p {
		if opt.k == option {
			return opt, true
		}
	}
	return Option{}, false
}

type Option struct {
	raw, k, v string
}

func (p Option) String() string { return p.raw }
func (p Option) Key() string    { return p.k }
func (p Option) Value() string  { return p.v }

func ParseOptions(tag string, optionSep, kvSep string) Options {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil
	}

	var options []Option
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
