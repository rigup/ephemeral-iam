// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

// See https://github.com/rhaidiz/broxy/modules/coreproxy/coreproxy.go
func setCa(caCertFile, caKeyFile string) error {
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return errorsutil.New(fmt.Sprintf("Failed to read CA certificate file %s", caCertFile), err)
	}
	caKey, err := ioutil.ReadFile(caKeyFile)
	if err != nil {
		return errorsutil.New(fmt.Sprintf("Failed to read CA certificate key file %s", caCertFile), err)
	}

	ca, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return errorsutil.New("Failed to parse X509 public/private key pair", err)
	}

	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return errorsutil.New("Failed to parse x509 certificate", err)
	}

	goproxy.GoproxyCa = ca
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: tlsConfigFromCA(&ca)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: tlsConfigFromCA(&ca)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: tlsConfigFromCA(&ca)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: tlsConfigFromCA(&ca)}
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
			cert, err = signHost(ca, hostname)
			if err != nil {
				return nil, err
			}
			setCachedCert(hostname, port, cert)
		}

		config := tls.Config{
			Certificates: []tls.Certificate{*cert},
			MinVersion:   tls.VersionTLS12,
		}

		return &config, nil
	}
}

// See https://github.com/rhaidiz/broxy/blob/master/core/cert.go
func signHost(ca *tls.Certificate, host string) (cert *tls.Certificate, err error) {
	var x509ca *x509.Certificate
	var template x509.Certificate

	if x509ca, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return
	}

	notBefore := time.Now()
	aYear := time.Until(notBefore.AddDate(1, 0, 0)) * time.Hour
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
			Locality:           []string{"Unknown"},
			Organization:       []string{"Unknown"},
			OrganizationalUnit: []string{"Unknown"},
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

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, x509ca, &certpriv.PublicKey, ca.PrivateKey)
	if err != nil {
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
