package downloader

import (
	"context"
	"io"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

type Downloader struct {
	url    string
	client *http.Client
}

func NewDownloader(url string) *Downloader {
	return &Downloader{
		url:    url,
		client: &http.Client{},
	}
}

func (d *Downloader) DownloadJSON(ctx context.Context) ([]byte, error) {
	txn := newrelic.FromContext(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.url, nil)
	if err != nil {
		return nil, err
	}

	// Create external segment
	segment := newrelic.StartExternalSegment(txn, req)
	defer segment.End()

	resp, err := d.client.Do(req)
	if err != nil {
		txn.NoticeError(err)
		return nil, err
	}
	defer resp.Body.Close()

	// Add response attributes
	txn.AddAttribute("http.status_code", resp.StatusCode)
	txn.AddAttribute("http.url", d.url)
	txn.AddAttribute("http.method", "GET")

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		txn.NoticeError(err)
		return nil, err
	}

	txn.AddAttribute("response.size_bytes", len(data))
	return data, nil
}
