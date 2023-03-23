package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"time"

	zlog "github.com/rs/zerolog/log"
	bom "github.com/vault-thirteen/auxie/BOM"
	rs "github.com/vault-thirteen/auxie/ReaderSeeker"
	"github.com/vault-thirteen/errorz"
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

	// Decode the response.
	var respBody []byte
	var contentEncodingHasChanged bool
	respBody, contentEncodingHasChanged, err = s.decodeResponse(targetResponse)
	if err != nil {
		http.Error(w, "decoding error", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("")
		return
	}

	// Respond to the client.
	s.writeResponseHeaders(w, targetResponse, respBody, contentEncodingHasChanged)
}

func (s *Server) modifyRequest(req *http.Request) {
	req.RequestURI = ""
	req.Header.Del(header.HttpHeaderKeepAlive)
	req.Header.Del(header.HttpHeaderConnection)
	req.Header.Add(header.HttpHeaderConnection, "close")
}

func (s *Server) decodeResponse(targetResponse *http.Response) (body []byte, contentEncodingHasChanged bool, err error) {
	defer func() {
		derr := targetResponse.Body.Close()
		if derr != nil {
			err = errorz.Combine(err, derr)
		}
	}()

	// bodyIsReady is set to true when HTTP body has been read.
	var bodyIsReady = false

	// Get the body.
	contentEncoding := targetResponse.Header.Get(header.HttpHeaderContentEncoding)
	if (contentEncoding == "gzip") || (contentEncoding == "x-gzip") {
		// Content is zipped.
		if s.parameters.MustDecodeGzip {
			var gzipReader *gzip.Reader
			gzipReader, err = gzip.NewReader(targetResponse.Body)
			if err != nil {
				return nil, contentEncodingHasChanged, err
			}
			defer func() {
				derr := gzipReader.Close()
				if derr != nil {
					err = errorz.Combine(err, derr)
				}
			}()

			body, err = io.ReadAll(gzipReader)
			if err != nil {
				return nil, contentEncodingHasChanged, err
			}

			// Save results.
			contentEncodingHasChanged = true
			bodyIsReady = true
		}
	}

	if !bodyIsReady {
		body, err = io.ReadAll(targetResponse.Body)
		if err != nil {
			return nil, contentEncodingHasChanged, err
		}

		// Save results.
		bodyIsReady = true
	}

	// Remove the BOM if needed.
	if s.parameters.MustRemoveBOM {
		body, err = s.removeBOM(body)
		if err != nil {
			return nil, contentEncodingHasChanged, err
		}
	}

	return body, contentEncodingHasChanged, nil
}

func (s *Server) writeResponseHeaders(w http.ResponseWriter, response *http.Response, respBody []byte, contentEncodingHasChanged bool) {
	if contentEncodingHasChanged {
		response.Header.Del("Content-Encoding")
	}

	for hdrName, lines := range response.Header {
		for _, line := range lines {
			w.Header().Add(hdrName, line)
		}
	}

	w.WriteHeader(response.StatusCode)

	_, err := w.Write(respBody)
	if err != nil {
		zlog.Error().Err(err).Msg("")
		return
	}
}

func (s *Server) removeBOM(data []byte) (newData []byte, err error) {
	rdr := bytes.NewReader(data)

	var possibleEncodings []bom.Encoding
	possibleEncodings, err = bom.GetEncoding(rdr, false)
	if err != nil {
		return nil, err
	}

	if len(possibleEncodings) == 0 {
		// Either no BOM is found or we do not know something.
		return data, nil
	}

	var encoding bom.Encoding
	if len(possibleEncodings) > 1 {
		// Guys, who created BOMs with similar bytes, you are not clever.
		encoding = possibleEncodings[0]
	} else {
		encoding = possibleEncodings[0]
	}

	var rdr2 rs.ReaderSeeker
	rdr2, err = bom.SkipBOMPrefix(rdr, encoding)
	if err != nil {
		return nil, err
	}

	newData, err = io.ReadAll(rdr2)
	if err != nil {
		return nil, err
	}

	return newData, nil
}
