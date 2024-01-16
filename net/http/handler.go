package http

import (
	log "github.com/sirupsen/logrus"
	"github.com/xvzc/SpoofDPI/dns"
	"github.com/xvzc/SpoofDPI/net"
)

func Handle(lConn *net.Conn, p *HttpRequest, timeout int) {
	log.Debug("[HTTP] Start")
	p.Tidy()

	ip, err := dns.Lookup(p.Domain())
	if err != nil {
		log.Error("[DNS] Error looking up for domain with ", p.Domain(), " ", err)
		lConn.Write([]byte(p.Version() + " 502 Bad Gateway\r\n\r\n"))
		return
	}

	log.Debug("[DNS] Found ", ip, " with ", p.Domain())

	// Create connection to server
	var port = "80"
	if p.Port() != "" {
		port = p.Port()
	}

	rConn, err := net.DialTCP("tcp", ip, port)
	if err != nil {
		log.Debug("[HTTP] ", err)
		return
	}

	defer func() {
		lConn.Close()
		log.Debug("[HTTP] Closing client Connection.. ", lConn.RemoteAddr())

		rConn.Close()
		log.Debug("[HTTP] Closing server Connection.. ", p.Domain(), " ", rConn.LocalAddr())
	}()

	log.Debug("[HTTP] New connection to the server ", p.Domain(), " ", rConn.LocalAddr())

	go rConn.Serve(lConn, "[HTTP]", lConn.RemoteAddr().String(), p.Domain(), timeout)

	_, err = rConn.Write(p.Raw())
	if err != nil {
		log.Debug("[HTTP] Error sending request to ", p.Domain(), err)
		return
	}

	log.Debug("[HTTP] Sent a request to ", p.Domain())

	lConn.Serve(rConn, "[HTTP]", lConn.RemoteAddr().String(), p.Domain(), timeout)

}
