package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"time"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
	keepAlivePeriod time.Duration
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(ln.keepAlivePeriod)
	return tc, nil
}

func tlsCert(name string, dur time.Duration) (tls.Certificate, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(0),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             now,
		NotAfter:              now.Add(dur),
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	cert, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv
	return outCert, nil
}

func oauthHandler(state string) (http.Handler, <-chan string) {
	authCodeC := make(chan string)
	mux := http.NewServeMux()
	mux.HandleFunc("/accept", func(rw http.ResponseWriter, r *http.Request) {
		actualState, err := url.QueryUnescape(r.URL.Query().Get("state"))
		if err != nil {
			http.Error(rw, fmt.Sprintf("invalid state: %v", err), http.StatusBadRequest)
			return
		}
		if string(actualState) != state {
			http.Error(rw, fmt.Sprintf("state mismatch"), http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		authCodeC <- code
		rw.Write([]byte("The test will proceed. You can safely close this tab."))
	})
	mux.HandleFunc("/policy", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Accept for the test to proceed."))
	})
	mux.HandleFunc("/decline", func(rw http.ResponseWriter, r *http.Request) {
		close(authCodeC)
		rw.Write([]byte("Accept for the test to proceed."))
	})
	return mux, authCodeC
}

func oauthServer(name, addr, state string) (serve func() error, teardown func() error, authCode <-chan string, err error) {
	cert, err := tlsCert(name, time.Hour)
	if err != nil {
		log.Fatal(err)
	}
	handler, authCodeC := oauthHandler(state)
	srv := &http.Server{Addr: addr, Handler: handler}
	cfg := &tls.Config{}
	cfg.NextProtos = []string{"http/1.1"}
	cfg.Certificates = make([]tls.Certificate, 1)
	cfg.Certificates[0] = cert
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, nil, err
	}
	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener), time.Minute}, cfg)
	return func() error { return srv.Serve(tlsListener) }, srv.Close, authCodeC, nil
}
