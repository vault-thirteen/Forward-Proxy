package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	ver "github.com/vault-thirteen/Versioneer"
)

func main() {
	showIntro()

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: http.HandlerFunc(handler),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func mustBeNoError(
	err error,
) {
	if err != nil {
		log.Fatal(err)
	}
}

func showIntro() {
	versioneer, err := ver.New()
	mustBeNoError(err)
	versioneer.ShowIntroText("")
	versioneer.ShowComponentsInfoText()
	fmt.Println()
}

func handler(w http.ResponseWriter, req *http.Request) {
	t1 := time.Now()

	switch req.Method {
	case http.MethodConnect:
		processHttpsRequest(w, req)
	default:
		processHttpRequest(w, req)
	}

	zlog.Debug().Msgf("time taken to serve '%v' is %v", req.URL.String(), time.Since(t1).String())
}

func processHttpRequest(w http.ResponseWriter, req *http.Request) {
	zlog.Debug().Msgf("http request to '%s'", req.URL.String())

	req.RequestURI = ""
	setClientHeaders(req.Header)

	client := &http.Client{}

	var targetResponse *http.Response
	var err error
	targetResponse, err = client.Do(req)
	if err != nil {
		http.Error(w, "client.do error", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("")
		return
	}

	setTargetHeaders(targetResponse.Header, w.Header())

	w.WriteHeader(targetResponse.StatusCode)

	// Copy body to ResponseWriter.
	_, err = io.Copy(w, targetResponse.Body)
	if err != nil {
		zlog.Error().Err(err).Msg("")
		return
	}
}

func processHttpsRequest(w http.ResponseWriter, req *http.Request) {
	zlog.Debug().Msgf("https request to '%s'", req.URL.String())

	// Establish a TCP connection with target.
	targetConn, err := net.Dial("tcp", req.URL.Host)
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

	clientConn, _, err := hjk.Hijack()
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
	go copyData(targetConn, clientConn, closer)
	go copyData(clientConn, targetConn, closer)
	<-closer
	<-closer
}

func copyData(target, client net.Conn, closer chan bool) {
	defer func() {
		closer <- true
	}()

	_, err := io.Copy(target, client)
	if err != nil {
		zlog.Error().Err(err).Msg("")
		return
	}
}
