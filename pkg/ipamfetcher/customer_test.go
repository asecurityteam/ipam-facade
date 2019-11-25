package ipamfetcher

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

type errReader struct {
	Error error
}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.Error
}

func TestNewDevice42CustomerFetcher(t *testing.T) {
	component := NewDevice42ClientComponent()
	config := &Device42ClientConfig{
		Endpoint: "https://localhost:443",
		Limit:    50,
		HTTP:     component.HTTP.Settings(),
	}
	client, _ := component.New(context.Background(), config)
	fetcher := NewDevice42CustomerFetcher(client)
	assert.Equal(t, "https://localhost:443/api/1.0/customers", fetcher.Endpoint.String())
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
	endpoint, _ := url.Parse("http://locaEndpoint")

	c := &Device42CustomerFetcher{Endpoint: endpoint, Client: &http.Client{Transport: mockRT}}

	customers, err := c.FetchCustomers(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []domain.Customer{domain.Customer{ID: "1", ResourceOwner: "foo@atlassian.com", BusinessUnit: "Security"}, domain.Customer{ID: "2", ResourceOwner: "bar@atlassian.com", BusinessUnit: "Bitbucket"}}, customers)
}

func TestFetchCustomersFallbackToName(t *testing.T) {
	// test that we fall back to the top level "name" when "Custom Fields" value for key=="Description" is empty
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"Customers": [{"id": 1, "contact_info": "foo@atlassian.com", "name": "BobTheBusinessUnit", "Custom Fields": [{"key": "Description", "value": ""}]}, {"id": 2, "contact_info": "bar@atlassian.com", "Custom Fields": [{"key": "Description", "value": "Bitbucket"}]}]}`))),
			StatusCode: http.StatusOK,
		},
		nil,
	)
	endpoint, _ := url.Parse("http://locaEndpoint")

	c := &Device42CustomerFetcher{Endpoint: endpoint, Client: &http.Client{Transport: mockRT}}

	customers, err := c.FetchCustomers(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []domain.Customer{domain.Customer{ID: "1", ResourceOwner: "foo@atlassian.com", BusinessUnit: "BobTheBusinessUnit"}, domain.Customer{ID: "2", ResourceOwner: "bar@atlassian.com", BusinessUnit: "Bitbucket"}}, customers)
}

func TestFetchCustomersNoContacts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"Customers": [{"id": 1, "contact_info": "contactinfo@atlassian.com", "Custom Fields": [{"key": "Description", "value": "Security"}]}]}`))),
			StatusCode: http.StatusOK,
		},
		nil,
	)
	endpoint, _ := url.Parse("http://locaEndpoint")

	c := &Device42CustomerFetcher{Endpoint: endpoint, Client: &http.Client{Transport: mockRT}}

	customers, err := c.FetchCustomers(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []domain.Customer{domain.Customer{ID: "1", ResourceOwner: "contactinfo@atlassian.com", BusinessUnit: "Security"}}, customers)
}

func TestFetchCustomersUseTeamLead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"Customers": [{"id": 1, "contact_info": "contactinfo@atlassian.com", "Custom Fields": [{"key": "Description", "value": "Security"}], "Contacts": [{"type": "Team Lead", "email": "teamlead@atlassian.com"}]}]}`))),
			StatusCode: http.StatusOK,
		},
		nil,
	)
	endpoint, _ := url.Parse("http://locaEndpoint")

	c := &Device42CustomerFetcher{Endpoint: endpoint, Client: &http.Client{Transport: mockRT}}

	customers, err := c.FetchCustomers(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []domain.Customer{domain.Customer{ID: "1", ResourceOwner: "teamlead@atlassian.com", BusinessUnit: "Security"}}, customers)
}

func TestFetchCustomersUseAdministrative(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"Customers": [{"id": 1, "contact_info": "contactinfo@atlassian.com", "Custom Fields": [{"key": "Description", "value": "Security"}], "Contacts": [{"type": "Administrative", "email": "administrative@atlassian.com"}, {"type": "SRE", "email": "sre@atlassian.com"}]}]}`))),
			StatusCode: http.StatusOK,
		},
		nil,
	)
	endpoint, _ := url.Parse("http://locaEndpoint")

	c := &Device42CustomerFetcher{Endpoint: endpoint, Client: &http.Client{Transport: mockRT}}

	customers, err := c.FetchCustomers(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []domain.Customer{domain.Customer{ID: "1", ResourceOwner: "administrative@atlassian.com", BusinessUnit: "Security"}}, customers)
}

func TestFetchCustomersUseSRE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"Customers": [{"id": 1, "contact_info": "contactinfo@atlassian.com", "Custom Fields": [{"key": "Description", "value": "Security"}], "Contacts": [{"type": "Technical", "email": "technical@atlassian.com"}, {"type": "SRE", "email": "sre@atlassian.com"}]}]}`))),
			StatusCode: http.StatusOK,
		},
		nil,
	)
	endpoint, _ := url.Parse("http://locaEndpoint")

	c := &Device42CustomerFetcher{Endpoint: endpoint, Client: &http.Client{Transport: mockRT}}

	customers, err := c.FetchCustomers(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []domain.Customer{domain.Customer{ID: "1", ResourceOwner: "sre@atlassian.com", BusinessUnit: "Security"}}, customers)
}

func TestFetchCustomersUseTechnical(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(
		&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"Customers": [{"id": 1, "contact_info": "contactinfo@atlassian.com", "Custom Fields": [{"key": "Description", "value": "Security"}], "Contacts": [{"type": "Technical", "email": "technical@atlassian.com"}]}]}`))),
			StatusCode: http.StatusOK,
		},
		nil,
	)
	endpoint, _ := url.Parse("http://locaEndpoint")

	c := &Device42CustomerFetcher{Endpoint: endpoint, Client: &http.Client{Transport: mockRT}}

	customers, err := c.FetchCustomers(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []domain.Customer{domain.Customer{ID: "1", ResourceOwner: "technical@atlassian.com", BusinessUnit: "Security"}}, customers)
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
			endpoint, _ := url.Parse("http://localhost")

			c := &Device42CustomerFetcher{Endpoint: endpoint, Client: &http.Client{Transport: mockRT}}

			_, err := c.FetchCustomers(context.Background())
			assert.NotNil(t, err)
		})
	}
}
