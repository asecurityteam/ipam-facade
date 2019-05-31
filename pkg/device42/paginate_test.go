package device42

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testPayload struct {
	TotalCount int    `json:"total_count"`
	Value      string `json:"value"`
}

func TestBatchPagedRequestSingleRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRT := NewMockRoundTripper(ctrl)
	endpoint, _ := url.Parse("http://localhost")
	jsonBody, _ := json.Marshal(testPayload{
		Value:      "foo",
		TotalCount: 1,
	})
	response := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(jsonBody)),
		StatusCode: http.StatusOK,
	}
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(response, nil)

	p := &Device42Paginator{
		Client:   &http.Client{Transport: mockRT},
		Endpoint: endpoint,
		Limit:    1,
	}

	res, err := p.BatchPagedRequests(context.Background())
	assert.ElementsMatch(t, res, [][]byte{[]byte(`{"total_count":1,"value":"foo"}`)})
	assert.Nil(t, err)
}

func TestBatchPagedRequestSingleRequestErrors(t *testing.T) {
	tc := []struct {
		name          string
		response      *http.Response
		responseError error
	}{
		{
			"initial request error",
			nil,
			fmt.Errorf("initial request error"),
		},
		{
			"unmarshal error",
			&http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{notjson}"))),
				StatusCode: http.StatusOK,
			},
			nil,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRT := NewMockRoundTripper(ctrl)
			mockRT.EXPECT().RoundTrip(gomock.Any()).Return(test.response, test.responseError)
			endpoint, _ := url.Parse("http://localhost")
			p := &Device42Paginator{
				Client:   &http.Client{Transport: mockRT},
				Endpoint: endpoint,
				Limit:    1,
			}
			_, err := p.BatchPagedRequests(context.Background())
			assert.NotNil(tt, err)
		})
	}
}

func TestBatchPagedRequestMultipleRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRT := NewMockRoundTripper(ctrl)
	endpoint, _ := url.Parse("http://localhost")
	jsonBody1, _ := json.Marshal(testPayload{
		Value:      "foo",
		TotalCount: 2,
	})
	response1 := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(jsonBody1)),
		StatusCode: http.StatusOK,
	}

	jsonBody2, _ := json.Marshal(testPayload{
		Value:      "bar",
		TotalCount: 2,
	})
	response2 := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(jsonBody2)),
		StatusCode: http.StatusOK,
	}
	gomock.InOrder(
		mockRT.EXPECT().RoundTrip(gomock.Any()).Return(response1, nil),
		mockRT.EXPECT().RoundTrip(gomock.Any()).Return(response2, nil),
	)

	p := &Device42Paginator{
		Client:   &http.Client{Transport: mockRT},
		Endpoint: endpoint,
		Limit:    1,
	}

	res, err := p.BatchPagedRequests(context.Background())
	assert.ElementsMatch(t, res, [][]byte{[]byte(`{"total_count":2,"value":"foo"}`), []byte(`{"total_count":2,"value":"bar"}`)})
	assert.Nil(t, err)
}

func TestBatchPagedRequestMultipleRequestsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRT := NewMockRoundTripper(ctrl)
	endpoint, _ := url.Parse("http://localhost")
	jsonBody1, _ := json.Marshal(testPayload{
		Value:      "foo",
		TotalCount: 2,
	})
	response1 := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(jsonBody1)),
		StatusCode: http.StatusOK,
	}

	gomock.InOrder(
		mockRT.EXPECT().RoundTrip(gomock.Any()).Return(response1, nil),
		mockRT.EXPECT().RoundTrip(gomock.Any()).Return(nil, fmt.Errorf("paged response err")),
	)

	p := &Device42Paginator{
		Client:   &http.Client{Transport: mockRT},
		Endpoint: endpoint,
		Limit:    1,
	}

	res, err := p.BatchPagedRequests(context.Background())
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

type errReader struct {
	Error error
}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.Error
}

func TestMakePagedRequestReadError(t *testing.T) {
	tc := []struct {
		name             string
		body             io.ReadCloser
		expectedError    bool
		expectedResponse []byte
	}{
		{
			"read error",
			ioutil.NopCloser(&errReader{Error: errors.New("ioutil.ReadAll error")}),
			true,
			nil,
		},
		{
			"success",
			ioutil.NopCloser(bytes.NewBufferString(`{"total_count": 1, "foo": "bar"}`)),
			false,
			[]byte(`{"total_count": 1, "foo": "bar"}`),
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRT := NewMockRoundTripper(ctrl)
			mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
				Body:       test.body,
				StatusCode: http.StatusOK,
			}, nil)

			endpoint, _ := url.Parse("http://localhost")
			p := &Device42Paginator{
				Client:   &http.Client{Transport: mockRT},
				Endpoint: endpoint,
				Limit:    1,
			}

			res, err := p.makePagedRequest(context.Background(), 0)

			if test.expectedError {
				assert.NotNil(tt, err)
			} else {
				assert.Nil(tt, err)
				assert.Equal(tt, test.expectedResponse, res)
			}
		})
	}
}
