package string

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var isMn = runes.Predicate(func(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
})

var unaccentTransformer = transform.Chain(norm.NFD, runes.Remove(isMn), norm.NFC)

// UnaccentString removes accents from a string.
func UnaccentString(s string) (string, error) {
	result, _, err := transform.String(unaccentTransformer, s)
	if err != nil {
		return "", fmt.Errorf("failed to unaccent string: %w", err)
	}

	return result, nil
}

// UnaccentReader removes accents from a reader.
func UnaccentReader(r io.Reader) io.Reader {
	return transform.NewReader(r, unaccentTransformer)
}

// NormalizeFileName returns a normalized file name from a string.
// Unaccent the name and replace all non asci characters to underscore.
func NormalizeFileName(name string) (string, error) {
	if name == "" {
		return "", errors.New("empty name")
	}

	name, err := UnaccentString(name)
	if err != nil {
		return "", err
	}
	ext := filepath.Ext(name)
	nameSExt := strings.TrimSuffix(name, ext)
	if nameSExt != "" {
		nameSExt = strings.Trim(nameSExt, " ")
		nameSExt = filepath.Clean(strings.ReplaceAll(nameSExt, "..", ""))
		nameSExt = strings.TrimLeft(nameSExt, "/")
		nameSExt = strings.TrimRight(nameSExt, "/")
		nameR := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
		nameSExt = nameR.ReplaceAllString(nameSExt, "-")
		nameR = regexp.MustCompile(`-{2,}`)
		nameSExt = nameR.ReplaceAllString(nameSExt, "-")
	}

	return nameSExt + ext, nil
}
