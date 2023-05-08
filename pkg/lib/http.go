package lib

import (
	"net"
	"net/http"
	"time"
)

func NewHttpClient(appConf *AppConfig) *http.Client {
	var httpTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 15 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   50,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	cl := &http.Client{
		Transport: httpTransport,
		Timeout:   appConf.HTTPClient.Timeout,
	}

	return cl
}
