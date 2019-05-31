package device42

import (
	"context"
	"errors"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFetchIPAMData(t *testing.T) {
	tc := []struct {
		name                  string
		subnetFetcherResult   []domain.Subnet
		subnetFetcherErr      error
		deviceFetcherResult   []domain.Device
		deviceFetcherErr      error
		customerFetcherResult []domain.Customer
		customerFetcherErr    error
		expectError           bool
	}{
		{
			"success",
			[]domain.Subnet{domain.Subnet{ID: "1"}},
			nil,
			[]domain.Device{domain.Device{ID: "1"}},
			nil,
			[]domain.Customer{domain.Customer{ID: "1"}},
			nil,
			false,
		},
		{
			"subnet fetch err",
			nil,
			errors.New("subnet fetch error"),
			[]domain.Device{domain.Device{ID: "1"}},
			nil,
			[]domain.Customer{domain.Customer{ID: "1"}},
			nil,
			true,
		},
		{
			"device fetch err",
			[]domain.Subnet{domain.Subnet{ID: "1"}},
			nil,
			nil,
			errors.New("device fetch error"),
			[]domain.Customer{domain.Customer{ID: "1"}},
			nil,
			true,
		},
		{
			"customer fetch err",
			[]domain.Subnet{domain.Subnet{ID: "1"}},
			nil,
			[]domain.Device{domain.Device{ID: "1"}},
			nil,
			nil,
			errors.New("device fetch error"),
			true,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockSubnetFetcher := NewMockSubnetFetcher(ctrl)
			mockSubnetFetcher.EXPECT().FetchSubnets(gomock.Any()).Return(test.subnetFetcherResult, test.subnetFetcherErr)
			mockDeviceFetcher := NewMockDeviceFetcher(ctrl)
			mockDeviceFetcher.EXPECT().FetchDevices(gomock.Any()).Return(test.deviceFetcherResult, test.deviceFetcherErr)
			mockCustomerFetcher := NewMockCustomerFetcher(ctrl)
			mockCustomerFetcher.EXPECT().FetchCustomers(gomock.Any()).Return(test.customerFetcherResult, test.customerFetcherErr)

			c := &Client{
				SubnetFetcher:   mockSubnetFetcher,
				DeviceFetcher:   mockDeviceFetcher,
				CustomerFetcher: mockCustomerFetcher,
			}

			_, err := c.FetchIPAMData(context.Background())
			assert.Equal(t, test.expectError, err != nil)
		})
	}
}
