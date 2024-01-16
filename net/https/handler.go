package https

import (
	log "github.com/sirupsen/logrus"
	"github.com/xvzc/SpoofDPI/dns"
	"github.com/xvzc/SpoofDPI/net"
	"github.com/xvzc/SpoofDPI/net/http"
)

func Handle(lConn *net.Conn, p *http.HttpRequest, timeout int) {
	log.Debug("[HTTPS] Start")
	ip, err := dns.Lookup(p.Domain())

	if err != nil {
		log.Error("[DNS] Error looking up for domain: ", p.Domain(), " ", err)
		lConn.Write([]byte(p.Version() + " 502 Bad Gateway\r\n\r\n"))
		return
	}

	log.Debug("[DNS] Found ", ip, " with ", p.Domain())

	// Create a connection to the requested server
	var port = "443"
	if p.Port() != "" {
		port = p.Port()
	}

	rConn, err := net.DialTCP("tcp4", ip, port)
	if err != nil {
		log.Debug("[HTTPS] ", err)
		return
	}

	defer func() {
		lConn.Close()
		log.Debug("[HTTPS] Closing client Connection.. ", lConn.RemoteAddr())

		rConn.Close()
		log.Debug("[HTTPS] Closing server Connection.. ", p.Domain(), " ", rConn.LocalAddr())
	}()

	log.Debug("[HTTPS] New connection to the server ", p.Domain(), " ", rConn.LocalAddr())

	_, err = lConn.Write([]byte(p.Version() + " 200 Connection Established\r\n\r\n"))
	if err != nil {
		log.Debug("[HTTPS] Error sending 200 Connection Established to the client", err)
		return
	}

	log.Debug("[HTTPS] Sent 200 Connection Estabalished to ", lConn.RemoteAddr())

	// Read client hello
	clientHello, err := lConn.ReadBytes()
	if err != nil {
		log.Debug("[HTTPS] Error reading client hello from the client", err)
		return
	}

	log.Debug("[HTTPS] Client sent hello ", len(clientHello), "bytes")

	// Generate a go routine that reads from the server

	pkt := ParseClientHello(clientHello)

	chunks := pkt.SplitInChunks()

	go rConn.Serve(lConn, "[HTTPS]", rConn.RemoteAddr().String(), p.Domain(), timeout)

	if _, err := rConn.WriteChunks(chunks); err != nil {
		log.Debug("[HTTPS] Error writing client hello to ", p.Domain(), err)
		return
	}

	lConn.Serve(rConn, "[HTTPS]", lConn.RemoteAddr().String(), p.Domain(), timeout)
}
