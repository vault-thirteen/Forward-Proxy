package server

import (
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"time"

	zlog "github.com/rs/zerolog/log"
	"github.com/vault-thirteen/errorz"
)

func (s *Server) router(w http.ResponseWriter, req *http.Request) {
	var t1 = time.Now()

	switch req.Method {
	case http.MethodConnect:
		s.processHttpsRequest(w, req)
	default:
		s.processHttpRequest(w, req)
	}

	zlog.Debug().Msgf("serve time of '%v' is %v ms",
		req.URL.String(), time.Since(t1).Milliseconds())
}

func (s *Server) processHttpsRequest(w http.ResponseWriter, req *http.Request) {
	zlog.Debug().Msgf("request to '%s'", req.URL.String())

	// Establish a TCP connection with target.
	targetConn, err := net.DialTimeout("tcp", req.URL.Host, s.parameters.targetConnectionTimeout)
	if err != nil {
		http.Error(w, "net.dial error", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("")
		return
	}

	defer func() {
		derr := targetConn.Close()
		if derr != nil {
			zlog.Error().Err(derr).Msg("")
			return
		}
	}()

	// Hijack the client's connection.
	hjk, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking is not supported", http.StatusInternalServerError)
		zlog.Error().Msg("hijacking is not supported")
		return
	}

	var clientConn net.Conn
	clientConn, _, err = hjk.Hijack()
	if err != nil {
		http.Error(w, "hijack error", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("")
		return
	}

	defer func() {
		derr := clientConn.Close()
		if derr != nil {
			zlog.Error().Err(derr).Msg("")
			return
		}
	}()

	// Accept the HTTPS upgrade.
	_, err = clientConn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
	if err != nil {
		http.Error(w, "write error", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("")
		return
	}

	closer := make(chan bool, 2)
	go s.copyData(targetConn, clientConn, &closer)
	go s.copyData(clientConn, targetConn, &closer)
	<-closer
	<-closer
}

func (s *Server) copyData(target, client net.Conn, closer *chan bool) {
	defer func() {
		*closer <- true
	}()

	//TODO: Add speed limiter.
	//TODO: Read s.mustStop flag.
	_, err := io.Copy(target, client)
	if err != nil {
		zlog.Error().Err(err).Msg("")
		return
	}
}

func (s *Server) processHttpRequest(w http.ResponseWriter, req *http.Request) {
	zlog.Debug().Msgf("http request to '%s'", req.URL.String())

	// Modify the original request.
	s.modifyRequest(req)

	client := &http.Client{}

	var targetResponse *http.Response
	var err error
	targetResponse, err = client.Do(req)
	if err != nil {
		http.Error(w, "client.do error", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("")
		return
	}

	// Decode the response.
	var respBody []byte
	respBody, err = s.decodeResponse(targetResponse)
	if err != nil {
		http.Error(w, "decoding error", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("")
		return
	}

	// Respond to the client.
	w.WriteHeader(targetResponse.StatusCode)
	s.writeResponseHeaders(targetResponse.Header, w)
	_, err = w.Write(respBody)
	if err != nil {
		zlog.Error().Err(err).Msg("")
		return
	}
}

func (s *Server) modifyRequest(req *http.Request) {
	req.RequestURI = ""
	req.Header.Del("Keep-Alive")
	req.Header.Del("Connection")
	req.Header.Add("Connection", "close")
}

// TODO: Remove BOM when it is present.
func (s *Server) decodeResponse(targetResponse *http.Response) (body []byte, err error) {
	contentEncoding := targetResponse.Header.Get("Content-Encoding")
	if (contentEncoding == "gzip") || (contentEncoding == "x-gzip") {
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(targetResponse.Body)
		if err != nil {
			return nil, err
		}
		defer func() {
			derr := gzipReader.Close()
			if derr != nil {
				err = errorz.Combine(err, derr)
			}
		}()

		body, err = io.ReadAll(gzipReader)
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	defer func() {
		derr := targetResponse.Body.Close()
		if derr != nil {
			err = errorz.Combine(err, derr)
		}
	}()

	body, err = io.ReadAll(targetResponse.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *Server) writeResponseHeaders(responseHeaders http.Header, w http.ResponseWriter) {
	for hdrName, lines := range responseHeaders {
		for _, line := range lines {
			w.Header().Add(hdrName, line)
		}
	}
}
