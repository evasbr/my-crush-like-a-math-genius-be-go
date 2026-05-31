// Package common provides cross-cutting utility helper functions
// that can be utilized across all layers of the application.
package common

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"evasbr/mclamg/exception"
	"io"
	"net"
	"net/http"
	"reflect"
	"time"
)

type HttpHeader struct {
	Key   string
	Value string
}

// ClientComponent is a generic helper struct to perform outbound REST API calls.
// It uses generic type [T] for request body and [E] for response body.
type ClientComponent[T any, E any] struct {
	HttpMethod     string
	UrlApi         string
	ConnectTimeout uint32
	ActiveTimeout  uint32
	Headers        []HttpHeader
	RequestBody    *T
	ResponseBody   *E
}

// defaultHTTPClient is a thread-safe singleton HTTP Client shared across the application.
// Reuses TCP connections (Connection Pool) and prevents operating system socket leaks.
var defaultHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 5 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // Dial connection timeout
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
}

// Execute performs the HTTP request based on the configurations within the ClientComponent.
// This function supports cancellation and dynamic timeouts via context propagation.
//
// RestClient Layer Usage Example:
//
//	type ProductDTO struct { Name string }
//	type ResponseDTO struct { Status string }
//
//	func (c *Client) CallAPI(ctx context.Context, req ProductDTO) ResponseDTO {
//	    var res ResponseDTO
//	    client := common.ClientComponent[ProductDTO, ResponseDTO]{
//	        HttpMethod: "POST",
//	        UrlApi: "https://api.gateway.com/products",
//	        RequestBody: &req,
//	        ResponseBody: &res,
//	        ActiveTimeout: 5000, // 5 seconds timeout
//	    }
//	    err := client.Execute(ctx)
//	    exception.PanicLogging(err)
//	    return res
//	}
func (c *ClientComponent[T, E]) Execute(ctx context.Context) error {
	// Set dynamic active timeout via context if specified (> 0)
	if c.ActiveTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.ActiveTimeout)*time.Millisecond)
		defer cancel()
	}

	var request *http.Request
	var response *http.Response
	var err error = nil

	// Marshal request body to JSON stream
	if reflect.ValueOf(c.RequestBody).IsZero() || c.RequestBody == nil {
		request, err = http.NewRequestWithContext(ctx, c.HttpMethod, c.UrlApi, nil)
		exception.PanicLogging(err)
	} else {
		requestBody, err := json.Marshal(c.RequestBody)
		exception.PanicLogging(err)

		// Log request payload
		Logger(ctx, "HttpClient").Info("Request Body ", string(requestBody))

		requestBodyByte := bytes.NewBuffer(requestBody)
		request, err = http.NewRequestWithContext(ctx, c.HttpMethod, c.UrlApi, requestBodyByte)
		exception.PanicLogging(err)
	}

	// Set HTTP Headers
	request.Header.Set("Content-Type", "application/json")
	for _, header := range c.Headers {
		request.Header.Set(header.Key, header.Value)
	}

	// Log request metadata before dispatching
	Logger(ctx, "HttpClient").Info("Request Url ", c.UrlApi)
	Logger(ctx, "HttpClient").Info("Request Method ", c.HttpMethod)
	Logger(ctx, "HttpClient").Info("Request Header ", request.Header)

	start := time.Now()

	// Execute HTTP request using Singleton Client
	response, err = defaultHTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close() // MANDATORY: Frees up connection to the connection pool

	elapsed := time.Since(start)

	// Read response stream
	responseBody, err := io.ReadAll(response.Body)
	exception.PanicLogging(err)

	// Unmarshal response body JSON to generic struct [E]
	err = json.Unmarshal(responseBody, &c.ResponseBody)
	exception.PanicLogging(err)

	// Log response details
	Logger(ctx, "HttpClient").Info("Received response for ", c.UrlApi, " in ", elapsed.Milliseconds(), " ms")
	Logger(ctx, "HttpClient").Info("Response Header ", response.Header)
	Logger(ctx, "HttpClient").Info("Response Http Status ", response.Status)
	Logger(ctx, "HttpClient").Info("Response Http Version ", response.Proto)
	Logger(ctx, "HttpClient").Info("Response Body ", string(responseBody))

	return nil
}
