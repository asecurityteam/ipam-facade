package ipamfetcher

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
		name                   string
		shouldCallFetchSubnets bool
		shouldCallFetchDevices bool
		subnetFetcherResult    []domain.Subnet
		subnetFetcherErr       error
		deviceFetcherResult    []domain.Device
		deviceFetcherErr       error
		customerFetcherResult  []domain.Customer
		customerFetcherErr     error
		expectError            bool
	}{
		{
			name:                   "success",
			shouldCallFetchSubnets: true,
			shouldCallFetchDevices: true,
			subnetFetcherResult:    []domain.Subnet{domain.Subnet{ID: "1"}},
			subnetFetcherErr:       nil,
			deviceFetcherResult:    []domain.Device{domain.Device{ID: "1"}},
			deviceFetcherErr:       nil,
			customerFetcherResult:  []domain.Customer{domain.Customer{ID: "1"}},
			customerFetcherErr:     nil,
			expectError:            false,
		},
		{
			name:                   "customer fetch err",
			shouldCallFetchSubnets: false,
			shouldCallFetchDevices: false,
			subnetFetcherResult:    []domain.Subnet{domain.Subnet{ID: "1"}},
			subnetFetcherErr:       nil,
			deviceFetcherResult:    []domain.Device{domain.Device{ID: "1"}},
			deviceFetcherErr:       nil,
			customerFetcherResult:  nil,
			customerFetcherErr:     errors.New("device fetch error"),
			expectError:            true,
		},
		{
			name:                   "subnet fetch err",
			shouldCallFetchSubnets: true,
			shouldCallFetchDevices: false,
			subnetFetcherResult:    nil,
			subnetFetcherErr:       errors.New("subnet fetch error"),
			deviceFetcherResult:    []domain.Device{domain.Device{ID: "1"}},
			deviceFetcherErr:       nil,
			customerFetcherResult:  []domain.Customer{domain.Customer{ID: "1"}},
			customerFetcherErr:     nil,
			expectError:            true,
		},
		{
			name:                   "device fetch err",
			shouldCallFetchSubnets: true,
			shouldCallFetchDevices: true,
			subnetFetcherResult:    []domain.Subnet{domain.Subnet{ID: "1"}},
			subnetFetcherErr:       nil,
			deviceFetcherResult:    nil,
			deviceFetcherErr:       errors.New("device fetch error"),
			customerFetcherResult:  []domain.Customer{domain.Customer{ID: "1"}},
			customerFetcherErr:     nil,
			expectError:            true,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockCustomerFetcher := NewMockCustomerFetcher(ctrl)
			mockSubnetFetcher := NewMockSubnetFetcher(ctrl)
			mockDeviceFetcher := NewMockDeviceFetcher(ctrl)

			mockCustomerFetcher.EXPECT().FetchCustomers(gomock.Any()).Return(test.customerFetcherResult, test.customerFetcherErr)

			if test.shouldCallFetchSubnets {
				mockSubnetFetcher.EXPECT().FetchSubnets(gomock.Any()).Return(test.subnetFetcherResult, test.subnetFetcherErr)
			}

			if test.shouldCallFetchDevices {
				mockDeviceFetcher.EXPECT().FetchDevices(gomock.Any()).Return(test.deviceFetcherResult, test.deviceFetcherErr)
			}

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
