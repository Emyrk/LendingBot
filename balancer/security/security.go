package security

import (
	//"crypto/rand"
	//"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	//"crypto/x509/pkix"
	//"encoding/pem"
	//"errors"
	"fmt"
	//"math/big"
	//"net"
	//"time"
)

var _ = fmt.Println

func GetServerTLSConfig() (*tls.Config, error) {
	cert, err := GetServerTLSCertificate()
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	// pool.AddCert(cert.Leaf)
	pool.AppendCertsFromPEM([]byte(certString))

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientAuth:         tls.VerifyClientCertIfGiven,
		InsecureSkipVerify: false,
		RootCAs:            pool}, nil

}

func GetServerTLSCertificate() (tls.Certificate, error) {
	return tls.X509KeyPair([]byte(certString), []byte(keyString))
}

func GetClientTLSConfig() (*tls.Config, error) {
	pool := x509.NewCertPool()
	// pool.AddCert(cert.Leaf)
	pool.AppendCertsFromPEM([]byte(certString))

	return &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            pool}, nil
}

// func rootCertTmpl() (*x509.Certificate, error) {
// 	rootCertTmpl, err := CertTemplate()

// 	if err != nil {
// 		return nil, err
// 	}
// 	// describe what the certificate will be used for
// 	rootCertTmpl.IsCA = true
// 	rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
// 	rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
// 	rootCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

// 	return rootCertTmpl, nil
// }

// func CreateCert(template, parent *x509.Certificate, pub interface{}, parentPriv interface{}) (
// 	cert *x509.Certificate, certPEM []byte, err error) {

// 	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, parentPriv)
// 	if err != nil {
// 		return
// 	}
// 	// parse the resulting certificate so we can use it again
// 	cert, err = x509.ParseCertificate(certDER)
// 	if err != nil {
// 		return
// 	}
// 	// PEM encode the certificate (this is a standard TLS encoding)
// 	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
// 	certPEM = pem.EncodeToMemory(&b)
// 	return
// }

// // helper function to create a cert template with a serial number and other required fields
// func CertTemplate() (*x509.Certificate, error) {
// 	// generate a random serial number (a real cert authority would have some logic behind this)
// 	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
// 	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
// 	if err != nil {
// 		return nil, errors.New("failed to generate serial number: " + err.Error())
// 	}

// 	tmpl := x509.Certificate{
// 		SerialNumber:          serialNumber,
// 		Subject:               pkix.Name{Organization: []string{"HodlZone"}},
// 		SignatureAlgorithm:    x509.SHA256WithRSA,
// 		NotBefore:             time.Now(),
// 		NotAfter:              time.Now().Add(time.Hour), // valid for an hour
// 		BasicConstraintsValid: true,
// 	}
// 	return &tmpl, nil
// }

var certString = `
-----BEGIN CERTIFICATE-----
MIIF9jCCA96gAwIBAgIJAIDRmo9AETAlMA0GCSqGSIb3DQEBDQUAMFkxCzAJBgNV
BAYTAlVTMREwDwYDVQQIEwhEZWxhd2FyZTETMBEGA1UEBxMKV2lsbWluZ3RvbjER
MA8GA1UEChMISG9kbHpvbmUxDzANBgNVBAMTBlN0ZXZlbjAgFw0xNzA3MDQxNTQw
NTdaGA8yMTE3MDYxMDE1NDA1N1owWTELMAkGA1UEBhMCVVMxETAPBgNVBAgTCERl
bGF3YXJlMRMwEQYDVQQHEwpXaWxtaW5ndG9uMREwDwYDVQQKEwhIb2Rsem9uZTEP
MA0GA1UEAxMGU3RldmVuMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA
ubRWDyu2CsEWjzGvYQi3ueQQP9KtDpF+ku9R9xSRte1XJ7UmHJraA70EXzSxQg+1
Q0mV14+NxtT1TpV7WLcneNefg7J16APKOKhpGpamlRP17WVUyAiNiJkHGmUopOVE
rBZIkMGpQoqbnIMyLtkeH1rohSDPIg0jm+68vgk50AGiNGa448AnPyAt6QxWTWe0
Oas2juQ60elZ6j2VI+o40zzbatm9iJ/+FBoQG7iofnd9mYIJR3IJZSOZ6exIz4S9
zgLmq2eSavks5bCKn0EOMrzXt8oA9vtctdstpH/d9t3jCtML05yFbRQWi7gnvD5K
FNsCmyhPn3LRuUpS+hHBrHWZ7WUCpxaaMoEgPFWIueI2XcNs3rb8IfsmLLrn59DG
pXvTt1ruyPC5OSL5ShKjUPLtQErol73IVh4uHch8pZBYsVSAZH8HWfD8Nx92vDF1
HyjirbY6lz6i+mctgKv1cOv8ccX2hI3TgSmSMz5La4wobnKYnjHmb72DTyRZXEiF
W0tz9v4dkuvrwARpgKRnSCT/+PLPGN1XHHxmrpGw7idOUn3i/yI0/+iZV1GVwbgy
oGsRiTDb4ZPA3oWFtKqDxOSJvzMhBflf5MEzERo+BXwsJvHXpU/6OuowkkTXVmem
bjzgr45vbzfgEE76Fdb5Irr5Vs5jtWTMF1HERsCbsqkCAwEAAaOBvjCBuzAdBgNV
HQ4EFgQUcE4dqgs2S9AnEuCGqPmzRui9ItgwgYsGA1UdIwSBgzCBgIAUcE4dqgs2
S9AnEuCGqPmzRui9ItihXaRbMFkxCzAJBgNVBAYTAlVTMREwDwYDVQQIEwhEZWxh
d2FyZTETMBEGA1UEBxMKV2lsbWluZ3RvbjERMA8GA1UEChMISG9kbHpvbmUxDzAN
BgNVBAMTBlN0ZXZlboIJAIDRmo9AETAlMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcN
AQENBQADggIBAJ6XnfcsLzuSad3+jLTOOa6ecduUAmB62RMV2+fYH6i7Y+otpZSR
HsT60zkeI8Ecm/UWkPj+X3CviZ4ptsyw1SMXxF/bsG/aWZc67lR6yc7rCpKkAUfY
WFMo/dMbElrnZOxkTL2iQRHuzbVfo+Mft2fH6PE0ORLzaDSKX9JvV2QyWecdUgGy
L56ryCX/QOvHqM9ERDhBML2ehY7RUeXQjfbs3c8N0xuLB3dPA4Usr/cddFpbvsh/
SlAiPl51qlyYO2rnm3SyEHtJmPOyUYSIz0F27BZSIYcOBPPVS7Jt59C8oj0qolac
WYewWHwaAvVHipVjUPqlz3vpHzAqw2R3l0pGuX9IB8rLqs9Z+5+MAA/wN5icrHs+
zGFokmy9WTZ1DCAXYZzRjLfLB6pueSI3E5xXgt8rAmBwCA/MT7pH9Xl7pIv3Cn1B
7SKHNaOa54FecwyyEH2e0MD70zLUOCfKUZCcf04wW3Jp4tfFtxNuzmxJquQZh6bf
PttxCrarKsk68LgioZxwJT0c/43s6mXTOpWApPMp95ZkTpefm9LG6hrDtaGJmvGO
6LsWkJXmtAXACyqpXCv7wF6SOJvJymHYpdyMd8wYPHX2EmSkDFXyCmVmBytgWQLS
A8k1WnqXla4Ke5vo5/bBEM+eBqaqfsrvvqzpi/exlV3/kFii7oZZn+gH
-----END CERTIFICATE-----
`

var keyString = `
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----

`
