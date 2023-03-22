package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	zlog "github.com/rs/zerolog/log"
)

type Server struct {
	parameters *Parameters

	// HTTP(S) server.
	listenDsn  string
	httpServer *http.Server

	// Channel for an external controller. When a message comes from this
	// channel, a controller must stop this server. The server does not stop
	// itself.
	mustBeStopped chan bool

	// Internal control structures.
	subRoutines *sync.WaitGroup
	mustStop    *atomic.Bool
	httpErrors  chan error
}

func NewServer(p *Parameters) (srv *Server, err error) {
	srv = &Server{
		parameters:    p,
		listenDsn:     net.JoinHostPort(p.Host, strconv.Itoa(int(p.Port))),
		httpServer:    nil, // See below.
		mustBeStopped: make(chan bool, 2),
		subRoutines:   new(sync.WaitGroup),
		mustStop:      new(atomic.Bool),
		httpErrors:    make(chan error, 8),
	}

	srv.httpServer = &http.Server{
		Addr:    srv.listenDsn,
		Handler: http.HandlerFunc(srv.router),
	}

	return srv, nil
}

func (s *Server) GetListenDsn() (dsn string) {
	return s.listenDsn
}

func (s *Server) GetStopChannel() *chan bool {
	return &s.mustBeStopped
}

func (s *Server) Start() (err error) {
	s.startHttpServer()

	s.subRoutines.Add(1)
	go s.listenForHttpErrors()

	return nil
}

func (s *Server) Stop() (err error) {
	s.mustStop.Store(true)

	ctx, cf := context.WithTimeout(context.Background(), time.Minute)
	defer cf()
	err = s.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	close(s.httpErrors)

	s.subRoutines.Wait()

	return nil
}

func (s *Server) startHttpServer() {
	go func() {
		var listenError error
		listenError = s.httpServer.ListenAndServe()
		if (listenError != nil) && (listenError != http.ErrServerClosed) {
			s.httpErrors <- listenError
		}
	}()
}

func (s *Server) listenForHttpErrors() {
	defer s.subRoutines.Done()

	for err := range s.httpErrors {
		zlog.Error().Msg("HTTP Server error: " + err.Error())
		s.mustBeStopped <- true
	}

	log.Println("HTTP error listener has stopped.")
}
