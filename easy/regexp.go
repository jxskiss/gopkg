package easy

import "regexp"

// MatchGroups returns the matched named capturing groups.
// A returned value of nil indicates no match.
func MatchGroups(re *regexp.Regexp, str []byte) map[string][]byte {
	match := re.FindSubmatch(str)
	if len(match) == 0 {
		return nil
	}

	out := make(map[string][]byte, len(re.SubexpNames())-1)
	for i, key := range re.SubexpNames() {
		if i > 0 && key != "" {
			out[key] = match[i]
		}
	}
	return out
}

// MatchStringGroups returns the matched named capturing groups.
// A returned value of nil indicates no match.
func MatchStringGroups(re *regexp.Regexp, str string) map[string]string {
	match := re.FindStringSubmatch(str)
	if len(match) == 0 {
		return nil
	}

	out := make(map[string]string, len(re.SubexpNames())-1)
	for i, key := range re.SubexpNames() {
		if i > 0 && key != "" {
			out[key] = match[i]
		}
	}
	return out
}
