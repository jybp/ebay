package integration

// From https://gist.github.com/shivakar/cd52b5594d4912fbeb46

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"
)

// From https://golang.org/src/net/http/server.go
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
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

func setupTLS() *http.ServeMux {
	cert, err := tlsCert("eSniper", time.Hour)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":52125", Handler: mux}
	cfg := &tls.Config{}
	cfg.NextProtos = []string{"http/1.1"}
	cfg.Certificates = make([]tls.Certificate, 1)
	cfg.Certificates[0] = cert
	ln, err := net.Listen("tcp", ":52125")
	if err != nil {
		log.Fatal(err)
	}
	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener), time.Minute}, cfg)
	go func() {
		srv.Serve(tlsListener)
	}()
	return mux
}
