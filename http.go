package go_tronsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type HttpClient struct {
	client      *http.Client
	basePath    string
	lock        sync.Mutex
	postHeaders http.Header
	getHeaders  http.Header
	timeout     time.Duration
}

const (
	JsonContentType = "application/json"
)

func NewHttpClient(basePath string, timeoutSeconds int64) *HttpClient {
	client := &HttpClient{}
	client.getHeaders = make(http.Header)
	client.getHeaders.Set("accept", JsonContentType)
	client.postHeaders = make(http.Header)
	client.postHeaders.Set("accept", JsonContentType)
	client.postHeaders.Set("content-type", JsonContentType)
	client.client = new(http.Client)
	client.basePath = basePath
	if timeoutSeconds > 0 {
		client.timeout = time.Duration(timeoutSeconds) * time.Second
	} else {
		client.timeout = 10 * time.Second
	}
	return client
}

func (c *HttpClient) Close() {}

func (c *HttpClient) GetNextMaintenanceTime(cctx context.Context) (time.Time, error) {
	ccctx := cctx
	if ccctx == nil {
		ccctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ccctx, c.timeout)
	defer cancel()
	resp, err := c.doRequest(ctx, c.basePath+"/wallet/getnextmaintenancetime", false, nil)
	if err != nil {
		return time.Time{}, err
	}
	defer func() {
		_ = resp.Close()
	}()
	body := &struct {
		Num int64 `json:"num"`
	}{}
	if err = json.NewDecoder(resp).Decode(body); err != nil {
		return time.Time{}, err
	}
	if body.Num > 9999999999 {
		return time.UnixMilli(body.Num), nil
	} else {
		return time.Unix(body.Num, 0), nil
	}
}

func (c *HttpClient) doRequest(ctx context.Context, url string, post bool, msg interface{}) (io.ReadCloser, error) {
	var req *http.Request
	var err error
	if post {
		body, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(bytes.NewReader(body)))
		if err != nil {
			return nil, err
		}
		req.ContentLength = int64(len(body))
		req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
		// set headers
		c.lock.Lock()
		req.Header = c.postHeaders.Clone()
		c.lock.Unlock()
		setHeaders(req.Header, headersFromContext(ctx))
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		// set headers
		c.lock.Lock()
		req.Header = c.getHeaders.Clone()
		c.lock.Unlock()
		setHeaders(req.Header, headersFromContext(ctx))
	}

	// do request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var buf bytes.Buffer
		var body []byte
		if _, err := buf.ReadFrom(resp.Body); err == nil {
			body = buf.Bytes()
		}

		return nil, HTTPError{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}
	return resp.Body, nil
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (err HTTPError) Error() string {
	if len(err.Body) == 0 {
		return err.Status
	}
	return fmt.Sprintf("%v: %s", err.Status, err.Body)
}

type mdHeaderKey struct{}

// headersFromContext is used to extract http.Header from context.
func headersFromContext(ctx context.Context) http.Header {
	source, _ := ctx.Value(mdHeaderKey{}).(http.Header)
	return source
}

// setHeaders sets all headers from src in dst.
func setHeaders(dst http.Header, src http.Header) http.Header {
	for key, values := range src {
		dst[http.CanonicalHeaderKey(key)] = values
	}
	return dst
}
