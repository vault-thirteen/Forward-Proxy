package server

import (
	"net"
	"net/http"

	ipa "github.com/vault-thirteen/auxie/IPA"
)

func (s *Server) getClientIPAddress(req *http.Request) (ipaddr ipa.IPAddressV4, err error) {
	var clientHost string
	clientHost, _, err = net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return ipaddr, err
	}

	clientIPAddr := net.ParseIP(clientHost)
	if clientIPAddr == nil {
		return ipaddr, err
	}

	ipaddr, err = ipa.NewFromString(clientIPAddr.String())
	if err != nil {
		return ipaddr, err
	}

	return ipaddr, nil
}

func (s *Server) isIPAddressAllowed(ipaddr ipa.IPAddressV4) (ok bool) {
	if s.parameters.workMode.IsPrivate() {
		_, ok = s.parameters.workMode.WhiteList()[ipaddr]
		return ok
	}

	return true
}
