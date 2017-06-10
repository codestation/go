// +build vita

package x509

import (
	"io/ioutil"
	"os"
)

// Possible certificate files; stop after finding one.
var certFiles = []string{
	"vs0:data/external/cert/CA_LIST.cer",
}

func (c *Certificate) systemVerify(opts *VerifyOptions) (chains [][]*Certificate, err error) {
	return nil, nil
}

func loadSystemRoots() (*CertPool, error) {
	roots := NewCertPool()
	var bestErr error
	for _, file := range certFiles {
		data, err := ioutil.ReadFile(file)
		if err == nil {
			roots.AppendCertsFromPEM(data)
			return roots, nil
		}
		if bestErr == nil || (os.IsNotExist(bestErr) && !os.IsNotExist(err)) {
			bestErr = err
		}
	}
	return nil, bestErr
}
