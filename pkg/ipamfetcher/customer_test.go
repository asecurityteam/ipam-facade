package ipamfetcher

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type errReader struct {
	Error error
}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.Error
}

func TestFetchCustomers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"Customers": [{"id": 1, "contact_info": "foo@atlassian.com", "Custom Fields": [{"key": "Description", "value": "Security"}]}, {"id": 2, "contact_info": "bar@atlassian.com", "Custom Fields": [{"key": "Description", "value": "Bitbucket"}]}]}`))),
			StatusCode: http.StatusOK,
		},
		nil,
	)
	host, _ := url.Parse("http://localhost")

	c := &Device42CustomerFetcher{Host: host, Client: &http.Client{Transport: mockRT}}

	customers, err := c.FetchCustomers(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []domain.Customer{domain.Customer{ID: "1", ResourceOwner: "foo@atlassian.com", BusinessUnit: "Security"}, domain.Customer{ID: "2", ResourceOwner: "bar@atlassian.com", BusinessUnit: "Bitbucket"}}, customers)
}

func TestFetchCustomersRequestError(t *testing.T) {
	tc := []struct {
		name        string
		response    *http.Response
		responseErr error
	}{
		{
			"request error",
			nil,
			errors.New("request error"),
		},
		{
			"read error",
			&http.Response{Body: ioutil.NopCloser(&errReader{Error: errors.New("ioutil.ReadAll error")}), StatusCode: http.StatusOK},
			nil,
		},
		{
			"non-200 ok response",
			&http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("unexpected error"))), StatusCode: http.StatusBadRequest},
			nil,
		},
		{
			"unmarshal error",
			&http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("notjson"))), StatusCode: http.StatusOK},
			nil,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRT := NewMockRoundTripper(ctrl)
			mockRT.EXPECT().RoundTrip(gomock.Any()).Return(
				test.response,
				test.responseErr,
			)
			host, _ := url.Parse("http://localhost")

			c := &Device42CustomerFetcher{Host: host, Client: &http.Client{Transport: mockRT}}

			_, err := c.FetchCustomers(context.Background())
			assert.NotNil(t, err)
		})
	}
}
