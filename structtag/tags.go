package structtag

import (
	"strings"
)

type Options []Option

func (p Options) Get(option string) (string, bool) {
	for _, opt := range p {
		if opt.K == option {
			return opt.V, true
		}
	}
	return "", false
}

type Option struct {
	Value string
	K, V  string
}

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
		opt := Option{Value: optstr}
		if kvSep != "" {
			sepidx := strings.Index(optstr, kvSep)
			if sepidx < 0 {
				opt.K = optstr
			} else {
				opt.K = strings.TrimSpace(optstr[:sepidx])
				opt.V = strings.TrimSpace(optstr[sepidx+1:])
			}
		}
		options = append(options, opt)
	}
	return options
}
