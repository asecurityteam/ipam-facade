package ipamfetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPagedResponseUnmarshalJSON(t *testing.T) {
	payload := []byte(`{"total_count": 2, "limit": 1, "offset": 0, "value": "test"}`)
	var p PagedResponse
	err := json.Unmarshal(payload, &p)
	assert.Nil(t, err)
	assert.Equal(t, 2, p.TotalCount)
	assert.Equal(t, 1, p.Limit)
	assert.Equal(t, 0, p.Offset)
	assert.Equal(t, payload, p.Body)
}

func TestPagedResponseUnmarshalJSONErr(t *testing.T) {
	payload := []byte(`{"notvalidjson"}`)
	var p PagedResponse
	err := json.Unmarshal(payload, &p)
	assert.NotNil(t, err)
}

func TestDevice42PageIteratorCurrent(t *testing.T) {
	tc := []struct {
		name             string
		it               *Device42PageIterator
		expectedResponse PagedResponse
	}{
		{
			name:             "success",
			it:               &Device42PageIterator{offset: 1, totalCount: 10, currentPage: PagedResponse{Offset: 1}},
			expectedResponse: PagedResponse{Offset: 1},
		},
		{
			name:             "offset greater than totalCount",
			it:               &Device42PageIterator{offset: 11, totalCount: 10, currentPage: PagedResponse{Offset: 10}},
			expectedResponse: PagedResponse{},
		},
		{
			name:             "iterator error",
			it:               &Device42PageIterator{offset: 2, totalCount: 10, currentPage: PagedResponse{Offset: 2}, err: errors.New("error during last req")},
			expectedResponse: PagedResponse{},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			assert.Equal(tt, test.expectedResponse, test.it.Current())
		})
	}
}

func TestDevice42PageIteratorNext(t *testing.T) {
	iteratorLimit := 1
	tc := []struct {
		name                  string
		itOffset              int
		itTotalCount          int
		currentPageTotalCount int
		shouldCallFetchPage   bool
		pageFetchError        error
		expected              bool
		expectedOffset        int
	}{
		{
			name:                  "success",
			itOffset:              0,
			itTotalCount:          10,
			currentPageTotalCount: 10,
			shouldCallFetchPage:   true,
			pageFetchError:        nil,
			expected:              true,
			expectedOffset:        1,
		},
		{
			name:                  "PageFetcher error",
			itOffset:              0,
			itTotalCount:          10,
			currentPageTotalCount: 10,
			shouldCallFetchPage:   true,
			pageFetchError:        errors.New("request error"),
			expected:              false,
			expectedOffset:        0,
		},
		{
			name:                  "offset greater than totalCount",
			itOffset:              11,
			itTotalCount:          10,
			currentPageTotalCount: 10,
			shouldCallFetchPage:   false,
			pageFetchError:        nil,
			expected:              false,
			expectedOffset:        11,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(tt)
			defer ctrl.Finish()
			mockPageFetcher := NewMockPageFetcher(ctrl)
			if test.shouldCallFetchPage {
				mockPageFetcher.EXPECT().FetchPage(gomock.Any(), gomock.Any(), gomock.Any()).Return(PagedResponse{}, test.pageFetchError)
			}
			iterator := &Device42PageIterator{
				PageFetcher: mockPageFetcher,
				Limit:       iteratorLimit,
				currentPage: PagedResponse{TotalCount: test.currentPageTotalCount},
				offset:      test.itOffset,
				totalCount:  test.itTotalCount,
			}

			actual := iterator.Next()
			assert.Equal(tt, test.expected, actual)
			assert.Equal(tt, test.expectedOffset, iterator.offset)
		})
	}
}

func TestDevice42PageIteratorClose(t *testing.T) {
	p := &Device42PageIterator{err: errors.New("error")}
	err := p.Close()
	assert.NotNil(t, err)
}

func TestFetchPage(t *testing.T) {
	tc := []struct {
		name             string
		httpResponse     *http.Response
		httpError        error
		expectedError    bool
		expectedResponse PagedResponse
	}{
		{
			name:             "success",
			httpResponse:     &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{"total_count": 1, "offset": 0, "limit": 1, "foo": "bar"}`)), StatusCode: http.StatusOK},
			httpError:        nil,
			expectedError:    false,
			expectedResponse: PagedResponse{TotalCount: 1, Limit: 1, Offset: 0, Body: []byte(`{"total_count": 1, "offset": 0, "limit": 1, "foo": "bar"}`)},
		},
		{
			name:             "response error",
			httpResponse:     nil,
			httpError:        errors.New("req error"),
			expectedError:    true,
			expectedResponse: PagedResponse{},
		},
		{
			name:             "read error",
			httpResponse:     &http.Response{Body: ioutil.NopCloser(&errReader{Error: errors.New("ioutil.ReadAll error")}), StatusCode: http.StatusOK},
			httpError:        nil,
			expectedError:    true,
			expectedResponse: PagedResponse{},
		},
		{
			name:             "non-200 ok",
			httpResponse:     &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString("bad request")), StatusCode: http.StatusBadRequest},
			httpError:        nil,
			expectedError:    true,
			expectedResponse: PagedResponse{},
		},
		{
			name:             "unmarshal error",
			httpResponse:     &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{notjson}`)), StatusCode: http.StatusOK},
			httpError:        nil,
			expectedError:    true,
			expectedResponse: PagedResponse{},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRT := NewMockRoundTripper(ctrl)
			mockRT.EXPECT().RoundTrip(gomock.Any()).Return(test.httpResponse, test.httpError)

			endpoint, _ := url.Parse("http://localhost")
			p := &Device42PageFetcher{
				Client:   &http.Client{Transport: mockRT},
				Endpoint: endpoint,
			}

			res, err := p.FetchPage(context.Background(), 0, 1)

			if test.expectedError {
				assert.NotNil(tt, err)
			} else {
				assert.Nil(tt, err)
				assert.Equal(tt, test.expectedResponse, res)
			}
		})
	}
}
