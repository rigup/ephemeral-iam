package appconfig

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/emperror"
)

func GenerateCerts() {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	emperror.Panic(err)

	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	emperror.Panic(err)

	template := x509.Certificate{
		Version:      1,
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Locality:           []string{"Unknown"},
			Organization:       []string{"Unknown"},
			OrganizationalUnit: []string{"Unknown"},
			CommonName:         "Unknown",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	emperror.Panic(err)

	writeToFile(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}, "server.pem")
	writeToFile(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}, "server.key")
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func writeToFile(data *pem.Block, filename string) {
	filepath := filepath.Join(getConfigDir(), filename)
	fd, err := os.Create(filepath)
	emperror.Panic(err)
	defer fd.Close()

	pem.Encode(fd, data)
}
