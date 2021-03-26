package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
)

// See https://github.com/rhaidiz/broxy/modules/coreproxy/coreproxy.go
func setCa(caCertFile, caKeyFile string) error {
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return err
	}
	caKey, err := ioutil.ReadFile(caKeyFile)
	if err != nil {
		return err
	}

	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}

	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}

	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: tlsConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: tlsConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: tlsConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: tlsConfigFromCA(&goproxyCa)}
	return nil
}

// See https://github.com/rhaidiz/broxy/blob/master/core/cert.go
func tlsConfigFromCA(ca *tls.Certificate) func(host string, ctx *goproxy.ProxyCtx) (*tls.Config, error) {
	return func(host string, ctx *goproxy.ProxyCtx) (c *tls.Config, err error) {
		parts := strings.SplitN(host, ":", 2)
		hostname := parts[0]
		port := 443
		if len(parts) > 1 {
			port, err = strconv.Atoi(parts[1])
			if err != nil {
				port = 443
			}
		}

		cert := getCachedCert(hostname, port)
		if cert == nil {
			cert, err = signHost(ca, hostname, port)
			if err != nil {
				return nil, err
			}
			setCachedCert(hostname, port, cert)
		}

		config := tls.Config{
			Certificates: []tls.Certificate{*cert},
		}

		return &config, nil
	}
}

// See https://github.com/rhaidiz/broxy/blob/master/core/cert.go
func signHost(ca *tls.Certificate, host string, port int) (cert *tls.Certificate, err error) {
	var x509ca *x509.Certificate
	var template x509.Certificate

	if x509ca, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return
	}

	notBefore := time.Now()
	aYear := time.Duration(365) * time.Hour
	notAfter := notBefore.Add(aYear)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template = x509.Certificate{
		SerialNumber: serialNumber,
		Issuer:       x509ca.Subject,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Locality:           []string{""},
			Organization:       []string{"ephemeral-iam"},
			OrganizationalUnit: []string{"https://praetorian.com/"},
			CommonName:         "gcloud proxy cert",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	var certpriv *rsa.PrivateKey
	if certpriv, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return
	}

	var derBytes []byte
	if derBytes, err = x509.CreateCertificate(rand.Reader, &template, x509ca, &certpriv.PublicKey, ca.PrivateKey); err != nil {
		return
	}

	return &tls.Certificate{
		Certificate: [][]byte{derBytes, ca.Certificate[0]},
		PrivateKey:  certpriv,
	}, nil
}

func keyFor(domain string, port int) string {
	return fmt.Sprintf("%s:%d", domain, port)
}

func getCachedCert(domain string, port int) *tls.Certificate {
	certLock.Lock()
	defer certLock.Unlock()
	if cert, found := certCache[keyFor(domain, port)]; found {
		return cert
	}
	return nil
}

func setCachedCert(domain string, port int, cert *tls.Certificate) {
	certLock.Lock()
	defer certLock.Unlock()
	certCache[keyFor(domain, port)] = cert
}
