package proxy

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

// GenerateCerts creates the self signed TLS certificate for the HTTPS proxy
func GenerateCerts() error {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate random serial number limit for x509 cert: %v", err)
	}

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
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return fmt.Errorf("failed to create x509 Cert: %v", err)
	}

	if err := writeToFile(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}, "server.pem", 0o640); err != nil {
		return err
	}
	if err := writeToFile(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}, "server.key", 0o400); err != nil {
		return err
	}

	return nil
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

func writeToFile(data *pem.Block, filename string, perm os.FileMode) error {
	filepath := filepath.Join(appconfig.GetConfigDir(), filename)
	fd, err := os.Create(filepath)
	if err != nil {
		if os.IsPermission(err) {
			if err := os.Remove(filepath); err != nil {
				return fmt.Errorf("failed to update %s: %v", filepath, err)
			}
			return writeToFile(data, filename, perm)
		}
		return fmt.Errorf("failed to write file %s: %v", filepath, err)
	}

	defer fd.Close()

	if err := pem.Encode(fd, data); err != nil {
		return err
	}

	if err := os.Chmod(filepath, perm); err != nil {
		return err
	}
	return nil
}

func checkProxyCertificate() error {
	certFile := viper.GetString("authproxy.certfile")
	keyFile := viper.GetString("authproxy.keyfile")
	if certFile == "" || keyFile == "" {
		if keyFile == "" {
			util.Logger.Debug("Setting authproxy.keyfile")
			viper.Set("authproxy.keyfile", filepath.Join(appconfig.GetConfigDir(), "server.key"))
			keyFile = viper.GetString("authproxy.keyfile")
		}
		if certFile == "" {
			util.Logger.Debug("Setting authproxy.certfile")
			viper.Set("authproxy.certfile", filepath.Join(appconfig.GetConfigDir(), "server.pem"))
			certFile = viper.GetString("authproxy.certfile")
		}
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}

	_, certErr := os.Stat(certFile)
	_, keyErr := os.Stat(keyFile)
	certExists := !os.IsNotExist(certErr)
	keyExists := !os.IsNotExist(keyErr)

	if certExists != keyExists { // Check if only one of either the key or the cert exist
		util.Logger.Warn("Either the auth proxy cert or key is missing. Regenerating both")
		if err := GenerateCerts(); err != nil {
			return err
		}
	} else if !certExists { // Check if neither files exist
		if err := GenerateCerts(); err != nil {
			return err
		}
	}
	return nil
}
