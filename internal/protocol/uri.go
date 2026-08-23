package protocol

import (
	"errors"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ErrInvalidURI indicates a URI that does not meet Gopher-NG Core syntax.
var ErrInvalidURI = errors.New("invalid Gopher-NG URI")

// URI identifies a Gopher-NG Core resource.
type URI struct {
	Host     string
	Port     uint16
	Selector string
}

// ParseURI parses a Gopher-NG Core URI.
func ParseURI(raw string) (URI, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return URI{}, ErrInvalidURI
	}
	if !strings.EqualFold(u.Scheme, "gofer") || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(raw, "#") {
		return URI{}, ErrInvalidURI
	}

	host := u.Hostname()
	port := u.Port()
	if host == "" || port == "" || !decimal(port) || (strings.Contains(host, ":") && !strings.HasPrefix(u.Host, "[")) {
		return URI{}, ErrInvalidURI
	}
	if strings.HasPrefix(u.Host, "[") {
		address, err := netip.ParseAddr(host)
		if err != nil || !address.Is6() {
			return URI{}, ErrInvalidURI
		}
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return URI{}, ErrInvalidURI
	}

	// url.URL.Path has already decoded path escapes. Requiring it here avoids
	// treating an absent URI path as the root selector.
	if u.Path == "" || !utf8.ValidString(u.Path) {
		return URI{}, ErrInvalidURI
	}
	if err := ValidateSelector(u.Path); err != nil {
		return URI{}, ErrInvalidURI
	}

	return URI{Host: host, Port: uint16(portNumber), Selector: u.Path}, nil
}

func decimal(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
