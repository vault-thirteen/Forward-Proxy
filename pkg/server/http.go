package server

import (
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"time"

	zlog "github.com/rs/zerolog/log"
	"github.com/vault-thirteen/auxie/BOM/Reader"
	"github.com/vault-thirteen/header"
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

	// Apply processors to the data stream.
	var stream io.Reader
	var closers []io.Closer
	var mustClose bool
	var contentEncodingHasChanged bool
	closers = make([]io.Closer, 0)
	stream = targetResponse.Body

	if (stream != nil) && (stream != http.NoBody) {
		// 1. Gzip.
		stream, mustClose, contentEncodingHasChanged, err = s.processContentEncoding(targetResponse, stream)
		if err != nil {
			http.Error(w, "decoding error", http.StatusInternalServerError)
			zlog.Error().Err(err).Msg("")
			return
		}
		if mustClose {
			closers = append(closers, stream.(io.Closer))
		}

		// 2. BOM.
		stream, mustClose, err = s.processBOM(stream)
		if err != nil {
			http.Error(w, "decoding error", http.StatusInternalServerError)
			zlog.Error().Err(err).Msg("")
			return
		}
		if mustClose {
			closers = append(closers, stream.(io.Closer))
		}
	}

	defer func() {
		var derr error
		derr = targetResponse.Body.Close()
		if derr != nil {
			zlog.Error().Err(derr).Msg("")
		}

		for _, c := range closers {
			derr = c.Close()
			if derr != nil {
				zlog.Error().Err(derr).Msg("")
			}
		}
	}()

	if contentEncodingHasChanged {
		targetResponse.Header.Del(header.HttpHeaderContentEncoding)
	}

	// Respond to the client.
	err = s.writeResponse(w, stream, targetResponse)
	if err != nil {
		zlog.Error().Err(err).Msg("")
	}
}

func (s *Server) modifyRequest(req *http.Request) {
	req.RequestURI = ""
	req.Header.Del(header.HttpHeaderKeepAlive)
	req.Header.Del(header.HttpHeaderConnection)
	req.Header.Add(header.HttpHeaderConnection, "close")
}

func (s *Server) processContentEncoding(targetResponse *http.Response, inStream io.Reader) (outStream io.Reader, mustClose bool, contentEncodingHasChanged bool, err error) {
	contentEncoding := targetResponse.Header.Get(header.HttpHeaderContentEncoding)
	if (contentEncoding == "gzip") || (contentEncoding == "x-gzip") { // Content is Gzipped.
		if s.parameters.MustDecodeGzip { // We must decode the Gzip.
			var gzipReader *gzip.Reader
			gzipReader, err = gzip.NewReader(inStream)
			if err != nil {
				return inStream, false, false, err
			}
			return gzipReader, true, true, nil // Gzip decoder.
		}
	}

	return inStream, false, false, nil // No changes to the stream.
}

func (s *Server) processBOM(inStream io.Reader) (outStream io.Reader, mustClose bool, err error) {
	if s.parameters.MustRemoveBOM { // We must remove the BOM.
		var bomReader *reader.Reader
		bomReader, err = reader.NewReader(inStream, true)
		if err != nil {
			return inStream, false, err
		}
		return bomReader, true, nil // BOM remover.
	}

	return inStream, false, nil // No changes to the stream.
}

func (s *Server) writeResponse(w http.ResponseWriter, stream io.Reader, targetResponse *http.Response) (err error) {
	for hdrName, lines := range targetResponse.Header {
		for _, line := range lines {
			w.Header().Add(hdrName, line)
		}
	}

	w.WriteHeader(targetResponse.StatusCode)

	_, err = io.Copy(w, stream)
	if err != nil {
		return err
	}

	return nil
}
