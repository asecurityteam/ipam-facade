package ipamfetcher

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	d := &Device42ClientConfig{}
	assert.Equal(t, "Device42Client", d.Name())
}

func TestDefaultConfig(t *testing.T) {
	component := NewDevice42ClientComponent()
	config := component.Settings()
	client, err := component.New(context.Background(), config)
	zeroURL, _ := url.Parse("")
	assert.Equal(t, client.Endpoint, zeroURL)
	assert.Equal(t, client.Limit, 0)
	assert.NoError(t, err)
}

func TestBadEndpoint(t *testing.T) {
	component := NewDevice42ClientComponent()
	config := &Device42ClientConfig{
		Endpoint: "https://lo\\<calhost:443",
		HTTP:     component.HTTP.Settings(),
	}
	_, err := component.New(context.Background(), config)
	assert.Error(t, err)
}

func TestDevice42DependencyCheck(t *testing.T) {
	tests := []struct {
		name               string
		clientReturnStatus int
		expectedErr        bool
	}{
		{
			name:               "success",
			clientReturnStatus: http.StatusOK,
			expectedErr:        false,
		},
		{
			name:               "failure",
			clientReturnStatus: http.StatusTeapot,
			expectedErr:        true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(tt)
			mockRT := NewMockRoundTripper(ctrl)
			mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("🐖"))),
				StatusCode: test.clientReturnStatus,
			}, nil)
			clientURL, _ := url.Parse("http://localhost")
			client := Device42Client{
				Client:   &http.Client{Transport: mockRT},
				Endpoint: clientURL,
			}
			err := client.CheckDependencies(context.Background())
			assert.Equal(tt, test.expectedErr, err != nil)
		})
	}
}
