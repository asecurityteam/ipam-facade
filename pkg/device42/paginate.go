package device42

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

// Paginator is an interface that provides a concurrent fan-out to retrieve multiple
// response bodies from a paginated API
type Paginator interface {
	BatchPagedRequests(ctx context.Context) ([][]byte, error)
}

// Device42Paginator implements the Paginator interface for the Device42 paginated APIs
// See: https://api.device42.com/#api-get-limits-and-offsets
type Device42Paginator struct {
	Client   *http.Client
	Endpoint *url.URL
	Limit    int
}

type pagedPayload struct {
	TotalCount int `json:"total_count"`
}

// BatchPagedRequests performs a current fan-out to retrieve all responses from a Device42
// paginated API endpoint
func (p *Device42Paginator) BatchPagedRequests(ctx context.Context) ([][]byte, error) {
	initialResponse, err := p.makePagedRequest(ctx, 0)
	if err != nil {
		return nil, err
	}

	var pagedResponse pagedPayload
	err = json.Unmarshal(initialResponse, &pagedResponse)
	if err != nil {
		return nil, err
	}

	numRequests := int(math.Ceil(float64(pagedResponse.TotalCount) / float64(p.Limit)))
	responseChan := make(chan []byte, numRequests)
	errChan := make(chan error, 1)

	responseChan <- initialResponse
	for offset := p.Limit; offset <= (pagedResponse.TotalCount - p.Limit); offset = offset + p.Limit {
		go func(offset int) {
			resp, err := p.makePagedRequest(ctx, offset)
			if err != nil {
				errChan <- err
				return
			}
			responseChan <- resp
		}(offset)
	}

	responses := make([][]byte, 0, numRequests)
	for idx := 0; idx < numRequests; idx = idx + 1 {
		select {
		case resp := <-responseChan:
			responses = append(responses, resp)
		case err := <-errChan:
			return nil, err
		}
	}

	return responses, nil
}

func (p *Device42Paginator) makePagedRequest(ctx context.Context, offset int) ([]byte, error) {
	u, _ := url.Parse(p.Endpoint.String())
	q := u.Query()
	q.Set("limit", strconv.Itoa(p.Limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := p.Client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
