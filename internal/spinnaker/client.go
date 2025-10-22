package spinnaker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/spinnaker/spin/cmd/gateclient"
	"github.com/spinnaker/spin/config"
	gate "github.com/spinnaker/spin/gateapi"
)

const (
	SpinnakerClientInstanceKey = `SpinnakerClient`
)

//go:generate counterfeiter . Client
// Wrapper for the gate client to perform more specific actions.

type Client interface {
	GetPipelineExecution(string) (Execution, error)
}

type client struct {
	// Underlying gate client
	gc *gateclient.GatewayClient
	c  *http.Client
}

// NewClient Create new spinnaker gateway client with flag
func NewClient(ctx context.Context, c config.Config) (Client, error) {
	// Api client initialization.
	auth := c.Auth

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	httpclient := &http.Client{
		Jar: cookieJar,
	}

	if auth != nil && auth.Enabled && auth.X509 != nil {
		X509 := auth.X509
		httpclient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{},
		}

		if !X509.IsValid() {
			// Misconfigured.
			return nil, errors.New("incorrect x509 auth configuration.\nMust specify certPath/keyPath or cert/key pair")
		}

		if X509.Cert != "" && X509.Key != "" {
			certBytes := []byte(X509.Cert)
			keyBytes := []byte(X509.Key)

			cert, err := tls.X509KeyPair(certBytes, keyBytes)
			if err != nil {
				return nil, err
			}

			clientCertPool := x509.NewCertPool()
			clientCertPool.AppendCertsFromPEM(certBytes)

			httpclient.Transport.(*http.Transport).TLSClientConfig.MinVersion = tls.VersionTLS12
			httpclient.Transport.(*http.Transport).TLSClientConfig.Certificates = []tls.Certificate{cert}
		} else {
			// Misconfigured.
			return nil, errors.New("incorrect x509 auth configuration.\nMust specify certPath/keyPath or cert/key pair")
		}
	}

	gc := &gateclient.GatewayClient{
		Config:  c,
		Context: ctx,
	}
	cfg := &gate.Configuration{
		BasePath: gc.GateEndpoint(),
		// TODO version this correctly.
		UserAgent:  fmt.Sprintf("%s/%s", "spinkube", "unknown"),
		HTTPClient: httpclient,
	}
	gc.APIClient = gate.NewAPIClient(cfg)

	_, _, err = gc.VersionControllerApi.GetVersionUsingGET(gc.Context)
	if err != nil {
		fmt.Println("HERE: failed to get spinnaker version:", err)

		return nil, err
	}

	return &client{
		gc: gc,
		c:  httpclient,
	}, nil
}

func (c *client) GetPipelineExecution(id string) (Execution, error) {
	e := Execution{}

	resp, _, err := c.gc.PipelineControllerApi.GetPipelineUsingGET(c.gc.Context, id)
	if err != nil {
		return e, err
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return e, err
	}

	if err := json.Unmarshal(b, &e); err != nil {
		return e, err
	}

	e.gc = c.gc

	return e, nil
}
