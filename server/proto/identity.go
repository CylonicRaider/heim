package proto

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var validEmoji = map[string]string{}

var possibleEmoji = regexp.MustCompile(":[^\\s:]+?:")

const MaxNickLength = 36

const (
	ltrEmbed    = '\u202A'
	rtlEmbed    = '\u202B'
	ltrOverride = '\u202D'
	rtlOverride = '\u202E'
	ltrIsolate  = '\u2066'
	rtlIsolate  = '\u2067'
	fsIsolate   = '\u2068'

	bidiExplicitPop = '\u202C'
	bidiIsolatePop  = '\u2069'
)

type UserID string

func (uid UserID) String() string { return string(uid) }

func (uid UserID) Parse() (kind, id string) {
	parts := strings.SplitN(string(uid), ":", 2)
	if len(parts) < 2 {
		return "", string(uid)
	}
	return parts[0], parts[1]
}

// An Identity maps to a global persona. It may exist only in the context
// of a single Room. An Identity may be anonymous.
type Identity interface {
	ID() UserID
	Name() string
	ServerID() string
	View() IdentityView
}

type IdentityView struct {
	ID        UserID `json:"id"`         // the id of an agent or account
	Name      string `json:"name"`       // the name-in-use at the time this view was captured
	ServerID  string `json:"server_id"`  // the id of the server that captured this view
	ServerEra string `json:"server_era"` // the era of the server that captured this view
}

// LoadEmoji takes a json key-value object stored in the file at path and unmarshals it into
// the global validEmoji mapping.
func LoadEmoji(path string) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &validEmoji); err != nil {
		return err
	}
	return nil
}

// ApplyEmoji adds the given name-to-Unicode mappings into the emoji normalization table.
// For each mapping, the key is an emoji shortcode (without surrounding colons, like
// "astronaut") and the value is one of:
// - The empty string to delete any shortcode mapping;
// - A string of hex-encoded Unicode codepoints separated by hyphens denoting the
//   corresponding Unicode string (like "1f9d1-200d-1f680");
// - A string preceded by a tilde to denote a custom emoji (like "~euphorian-in-space").
func ApplyEmoji(emoji map[string]string) {
	if validEmoji == nil {
		validEmoji = map[string]string{}
	}
	for name, value := range emoji {
		if value == "" {
			delete(validEmoji, name)
		} else {
			validEmoji[name] = value
		}
	}
}

func nickLen(nick string) int {
	var result = 0
	for {
		m := possibleEmoji.FindStringIndex(nick)
		if m == nil {
			break
		}
		result += utf8.RuneCountInString(nick[:m[0]])

		v, ok := validEmoji[nick[m[0]+1 : m[1]-1]]
		if !ok {
			result += utf8.RuneCountInString(nick[m[0]:m[1]-1])
			nick = nick[m[1]-1:]
			continue
		}

		if v[0] == '~' {
			result += 1
		} else {
			result += 1 + strings.Count(v, "-")
		}
		nick = nick[m[1]:]
	}
	return result + utf8.RuneCountInString(nick)
}

// NormalizeNick validates and normalizes a proposed name from a user.
// If the proposed name is not valid, returns an error. Otherwise, returns
// the normalized form of the name. Normalization for a nick consists of:
//
// 1. Remove leading and trailing whitespace
// 2. Collapse all internal whitespace to single spaces
// 3. Close all unclosed bidi control codes
func NormalizeNick(name string) (string, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return "", ErrInvalidNick
	}
	normalized := strings.Join(strings.Fields(name), " ")
	if nickLen(normalized) > MaxNickLength {
		return "", ErrInvalidNick
	}
	return normalizeBidi(normalized), nil
}

// NormalizeNickAndEmoji validates and normalizes a proposed name from a user
// in a more extensive manner than NormalizeNick. In addition to
// NormalizeNick's steps, this does:
//
// 4. Replace emoji shortcodes with corresponding Unicode code points
func NormalizeNickAndEmoji(name string) (string, error) {
	result, err := NormalizeNick(name)
	if err != nil {
		return "", err
	}
	return normalizeEmoji(result), nil
}

// normalizeBidi attempts to prevent names from using bidi control codes to
// screw up our layout
func normalizeBidi(name string) string {
	bidiExplicitDepth := 0
	bidiIsolateDepth := 0

	for _, c := range name {
		switch c {
		case ltrEmbed, rtlEmbed, ltrOverride, rtlOverride:
			bidiExplicitDepth++
		case bidiExplicitPop:
			bidiExplicitDepth--
		case ltrIsolate, rtlIsolate, fsIsolate:
			bidiIsolateDepth++
		case bidiIsolatePop:
			bidiIsolateDepth--
		}
	}
	if bidiExplicitDepth+bidiIsolateDepth > 0 {
		pops := make([]byte,
			bidiExplicitDepth*utf8.RuneLen(bidiExplicitPop)+bidiIsolateDepth+utf8.RuneLen(bidiIsolatePop))
		i := 0
		for ; bidiExplicitDepth > 0; bidiExplicitDepth-- {
			i += utf8.EncodeRune(pops[i:], bidiExplicitPop)
		}
		for ; bidiIsolateDepth > 0; bidiIsolateDepth-- {
			i += utf8.EncodeRune(pops[i:], bidiIsolatePop)
		}
		return name + string(pops[:i])
	}
	return name
}

func codePointStringToUnicode(codepoints string) (string, bool) {
	ret := []rune{}
	for _, cp := range strings.Split(codepoints, "-") {
		cpi, err := strconv.ParseUint(cp, 16, 32)
		if err != nil {
			return "", false
		}
		ret = append(ret, rune(cpi))
	}
	return string(ret), true
}

// normalizeEmoji replaces emoji shortcodes with corresponding Unicode
// code points
func normalizeEmoji(nick string) string {
	sb := strings.Builder{}
	for {
		m := possibleEmoji.FindStringIndex(nick)
		if m == nil {
			break
		}
		shortcode := nick[m[0]+1 : m[1]-1]

		v, ok := validEmoji[shortcode]
		if !ok {
			sb.WriteString(nick[:m[1]-1])
			nick = nick[m[1]-1:]
			continue
		}

		sb.WriteString(nick[:m[0]])
		if v[0] == '~' {
			sb.WriteString(nick[m[0]:m[1]])
		} else if translated, ok := codePointStringToUnicode(v); !ok {
			panic("Invalid putative-Unicode emoji :" + shortcode + ": (" + v + ")")
		} else {
			sb.WriteString(translated)
		}
		nick = nick[m[1]:]
	}
	sb.WriteString(nick)
	return sb.String()
}
