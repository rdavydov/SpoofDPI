package http

import (
	"bufio"
	"net"
	"net/http"
	"strings"
)

var validMethod = map[string]struct{}{
	"DELETE":      {},
	"GET":         {},
	"HEAD":        {},
	"POST":        {},
	"PUT":         {},
	"CONNECT":     {},
	"OPTIONS":     {},
	"TRACE":       {},
	"COPY":        {},
	"LOCK":        {},
	"MKCOL":       {},
	"MOVE":        {},
	"PROPFIND":    {},
	"PROPPATCH":   {},
	"SEARCH":      {},
	"UNLOCK":      {},
	"BIND":        {},
	"REBIND":      {},
	"UNBIND":      {},
	"ACL":         {},
	"REPORT":      {},
	"MKACTIVITY":  {},
	"CHECKOUT":    {},
	"MERGE":       {},
	"M-SEARCH":    {},
	"NOTIFY":      {},
	"SUBSCRIBE":   {},
	"UNSUBSCRIBE": {},
	"PATCH":       {},
	"PURGE":       {},
	"MKCALENDAR":  {},
	"LINK":        {},
	"UNLINK":      {},
}

type HttpRequest struct {
	raw     []byte
	method  string
	domain  string
	port    string
	path    string
	version string
}

func ParseRequest(raw []byte) (*HttpRequest, error) {
	r := &HttpRequest{raw: raw}

	reader := bufio.NewReader(strings.NewReader(string(r.raw)))
	request, err := http.ReadRequest(reader)
	if err != nil {
		return nil, err
	}

	r.domain, r.port, err = net.SplitHostPort(request.Host)
	if err != nil {
		r.domain = request.Host
		r.port = ""
	}

	r.method = request.Method
	r.version = request.Proto
	r.path = request.URL.Path

	if request.URL.RawQuery != "" {
		r.path += "?" + request.URL.RawQuery
	}

	if request.URL.RawFragment != "" {
		r.path += "#" + request.URL.RawFragment
	}

	if r.path == "" {
		r.path = "/"
	}

	request.Body.Close()

	return r, nil
}

func (p *HttpRequest) Raw() []byte {
	return p.raw
}
func (p *HttpRequest) Method() string {
	return p.method
}

func (p *HttpRequest) Domain() string {
	return p.domain
}

func (p *HttpRequest) Port() string {
	return p.port
}

func (p *HttpRequest) Version() string {
	return p.version
}

func (p *HttpRequest) IsValidMethod() bool {
	if _, exists := validMethod[p.Method()]; exists {
		return true
	}

	return false
}

func (p *HttpRequest) IsConnectMethod() bool {
	return p.Method() == "CONNECT"
}

func (p *HttpRequest) Tidy() {
	s := string(p.raw)

	lines := strings.Split(s, "\r\n")

	lines[0] = p.method + " " + p.path + " " + p.version

	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "Proxy-Connection") {
			lines[i] = ""
		}
	}

	result := ""

	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}

		result += lines[i] + "\r\n"
	}

	result += "\r\n"

	p.raw = []byte(result)
}
