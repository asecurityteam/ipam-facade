package ipamfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// PagedResponse represents a standard structure for Device42 paginated response payloads
type PagedResponse struct {
	TotalCount int `json:"total_count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	Body       []byte
}

// UnmarshalJSON implements the Unmarshaler interface to extract known paginated fields from
// the response and stuff the entire response body into the Body struct for later
func (r *PagedResponse) UnmarshalJSON(data []byte) error {
	type Alias PagedResponse
	respAlias := &Alias{}
	err := json.Unmarshal(data, &respAlias)
	if err != nil {
		return err
	}
	r.TotalCount = respAlias.TotalCount
	r.Limit = respAlias.Limit
	r.Offset = respAlias.Offset
	r.Body = data
	return nil
}

// PageFetcher encapsulates the logic for requesting individual pages from a paged REST API
type PageFetcher interface {
	FetchPage(ctx context.Context, offset int, limit int) (PagedResponse, error)
}

// Iterator defines an interface for interacting with paginated APIs
type Iterator interface {
	// Next readies the next available element. It returns `false` when there
	// are no more items to iterate over.
	Next() bool
	// Current returns the current value of the iterator. This is non-empty
	// so long as Next() returned true.
	Current() PagedResponse
	// Close the iterator and fetch any error encountered. This returns nil
	// so long as the iterator closed cleanly.
	Close() error
}

// Device42PageIterator implements the iterator interface for paginated Device42 APIs
type Device42PageIterator struct {
	Context     context.Context
	PageFetcher PageFetcher
	Limit       int
	err         error
	currentPage PagedResponse
	offset      int
	totalCount  int
}

// Current returns the current response if there are no issues with the state of the iterator
func (it *Device42PageIterator) Current() PagedResponse {
	// Always have a guard against calls to `Current()` after the
	// iterator is complete. This is technically an invalid call
	// so "user beware" but this will prevent a panic condition.
	if it.offset > it.totalCount || it.err != nil {
		return PagedResponse{}
	}
	return it.currentPage
}

// Close returns the error from the iterator if any
func (it *Device42PageIterator) Close() error {
	return it.err
}

// Next fetches the next page from the API and makes necessary updates to iterator state
func (it *Device42PageIterator) Next() bool {
	if it.currentPage.TotalCount > 0 && it.offset >= it.totalCount {
		return false
	}

	nextPage, err := it.PageFetcher.FetchPage(it.Context, it.offset, it.Limit)
	if err != nil {
		it.err = err
		return false
	}

	it.currentPage = nextPage
	it.offset = it.offset + it.Limit
	it.totalCount = nextPage.TotalCount
	return true
}

// Device42PageFetcher implements the PageFetcher interface for Device42 APIs
type Device42PageFetcher struct {
	Client   *http.Client
	Endpoint *url.URL
}

// FetchPage makes a request to the Device42 with a given offset and limit
func (d *Device42PageFetcher) FetchPage(ctx context.Context, offset int, limit int) (PagedResponse, error) {
	u, _ := url.Parse(d.Endpoint.String())
	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := d.Client.Do(req.WithContext(ctx))
	if err != nil {
		return PagedResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return PagedResponse{}, fmt.Errorf("unexpected error from device42 api: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return PagedResponse{}, err
	}

	var pagedResponse PagedResponse
	if err := json.Unmarshal(body, &pagedResponse); err != nil {
		return PagedResponse{}, err
	}

	return pagedResponse, nil
}
