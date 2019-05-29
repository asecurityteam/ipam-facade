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

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetSubnetsSingleRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRT := NewMockRoundTripper(ctrl)
	host, _ := url.Parse("http://localhost")
	jsonBody, _ := json.Marshal(subnetResponse{
		Offset:     0,
		TotalCount: 1,
		Subnets:    []subnet{subnet{SubnetID: 1}},
	})
	response := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(jsonBody)),
		StatusCode: http.StatusOK,
	}
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(response, nil)

	c := &Client{
		Client: &http.Client{Transport: mockRT},
		Host:   host,
		Limit:  1,
	}

	subnets, err := c.GetSubnets(context.Background())
	assert.Equal(t, []domain.Subnet{domain.Subnet{ID: "1"}}, subnets)
	assert.Nil(t, err)
}

func TestGetSubnetsMultipleRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRT := NewMockRoundTripper(ctrl)
	host, _ := url.Parse("http://localhost")
	jsonBody1, _ := json.Marshal(subnetResponse{
		Limit:      1,
		Offset:     0,
		TotalCount: 2,
		Subnets:    []subnet{subnet{SubnetID: 1}},
	})
	response1 := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(jsonBody1)),
		StatusCode: http.StatusOK,
	}

	jsonBody2, _ := json.Marshal(subnetResponse{
		Limit:      1,
		Offset:     1,
		TotalCount: 2,
		Subnets:    []subnet{subnet{SubnetID: 2}},
	})
	response2 := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(jsonBody2)),
		StatusCode: http.StatusOK,
	}
	gomock.InOrder(
		mockRT.EXPECT().RoundTrip(gomock.Any()).Return(response1, nil),
		mockRT.EXPECT().RoundTrip(gomock.Any()).Return(response2, nil),
	)

	c := &Client{
		Client: &http.Client{Transport: mockRT},
		Host:   host,
		Limit:  1,
	}

	subnets, err := c.GetSubnets(context.Background())
	assert.Equal(t, []domain.Subnet{domain.Subnet{ID: "1"}, domain.Subnet{ID: "2"}}, subnets)
	assert.Nil(t, err)
}

func TestGetSubnetsMultipleRequestsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRT := NewMockRoundTripper(ctrl)
	host, _ := url.Parse("http://localhost")
	jsonBody1, _ := json.Marshal(subnetResponse{
		Limit:      1,
		Offset:     0,
		TotalCount: 2,
		Subnets:    []subnet{subnet{SubnetID: 1}},
	})
	response1 := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(jsonBody1)),
		StatusCode: http.StatusOK,
	}

	gomock.InOrder(
		mockRT.EXPECT().RoundTrip(gomock.Any()).Return(response1, nil),
		mockRT.EXPECT().RoundTrip(gomock.Any()).Return(nil, fmt.Errorf("error making second request")),
	)

	c := &Client{
		Client: &http.Client{Transport: mockRT},
		Host:   host,
		Limit:  1,
	}

	_, err := c.GetSubnets(context.Background())
	assert.NotNil(t, err)
}

func TestGetSubnetsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRT := NewMockRoundTripper(ctrl)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(nil, fmt.Errorf("req error"))

	host, _ := url.Parse("http://localhost")
	c := &Client{
		Client: &http.Client{Transport: mockRT},
		Host:   host,
		Limit:  1,
	}

	_, err := c.GetSubnets(context.Background())

	assert.NotNil(t, err)
}

type errReader struct {
	Error error
}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.Error
}

func TestMakePagedSubnetsRequestReadError(t *testing.T) {
	tc := []struct {
		name             string
		body             io.ReadCloser
		expectedError    bool
		expectedResponse subnetResponse
	}{
		{
			"read error",
			ioutil.NopCloser(&errReader{Error: errors.New("ioutil.ReadAll error")}),
			true,
			subnetResponse{},
		},
		{
			"unmarshal error",
			ioutil.NopCloser(bytes.NewBufferString("notjson")),
			true,
			subnetResponse{},
		},
		{
			"success",
			ioutil.NopCloser(bytes.NewBufferString(`{"limit": 1, "offset": 0, "total_count": 1, "subnets": [{"subnet_id": 1}]}`)),
			false,
			subnetResponse{Limit: 1, Offset: 0, TotalCount: 1, Subnets: []subnet{subnet{SubnetID: 1}}},
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

			host, _ := url.Parse("http://localhost")
			c := &Client{
				Client: &http.Client{Transport: mockRT},
				Host:   host,
				Limit:  1,
			}

			res, err := c.makePagedSubnetsRequest(context.Background(), 0)

			if test.expectedError {
				assert.NotNil(tt, err)
			} else {
				assert.Nil(tt, err)
				assert.Equal(tt, test.expectedResponse, res)
			}
		})
	}
}
