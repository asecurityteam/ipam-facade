package ipamfetcher

import (
	"context"
	"errors"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFetchDevices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPageFetcher := NewMockPageFetcher(ctrl)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 0, 1).Return(PagedResponse{TotalCount: 1, Offset: 0, Body: []byte(`{"offset": 0, "limit": 1, "total_count": 1, "ips": [{"ip": "192.168.1.1", "device_id": 1, "subnet_id": 1}]}`)}, nil)

	d := &Device42DeviceFetcher{
		Limit:       1,
		PageFetcher: mockPageFetcher,
	}

	assets, err := d.FetchDevices(context.Background())
	assert.ElementsMatch(t, []domain.Device{domain.Device{IP: "192.168.1.1", ID: "1", SubnetID: "1"}}, assets)
	assert.Nil(t, err)
}

func TestFetchDevicesMultiple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPageFetcher := NewMockPageFetcher(ctrl)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 0, 1).Return(PagedResponse{TotalCount: 3, Offset: 0, Body: []byte(`{"offset": 0, "limit": 1, "total_count": 3, "ips": [{"ip": "192.168.1.1", "device_id": 1, "subnet_id": 1}]}`)}, nil)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 1, 1).Return(PagedResponse{TotalCount: 3, Offset: 1, Body: []byte(`{"offset": 1, "limit": 1, "total_count": 3, "ips": [{"ip": "192.168.1.2", "device_id": 2, "subnet_id": 2}]}`)}, nil)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 2, 1).Return(PagedResponse{TotalCount: 3, Offset: 2, Body: []byte(`{"offset": 2, "limit": 1, "total_count": 3, "ips": [{"ip": "192.168.1.3", "device_id": 3, "subnet_id": 3}]}`)}, nil)

	d := &Device42DeviceFetcher{
		Limit:       1,
		PageFetcher: mockPageFetcher,
	}

	assets, err := d.FetchDevices(context.Background())
	assert.ElementsMatch(t, []domain.Device{domain.Device{IP: "192.168.1.1", ID: "1", SubnetID: "1"}, domain.Device{IP: "192.168.1.2", ID: "2", SubnetID: "2"}, domain.Device{IP: "192.168.1.3", ID: "3", SubnetID: "3"}}, assets)
	assert.Nil(t, err)
}

func TestFetchDevicesUnmarshalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPageFetcher := NewMockPageFetcher(ctrl)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 0, 1).Return(PagedResponse{TotalCount: 1, Offset: 0, Body: []byte(`notanip`)}, nil)

	d := &Device42DeviceFetcher{
		Limit:       1,
		PageFetcher: mockPageFetcher,
	}

	_, err := d.FetchDevices(context.Background())
	assert.NotNil(t, err)
}

func TestFetchDevicesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPageFetcher := NewMockPageFetcher(ctrl)
	mockPageFetcher.EXPECT().FetchPage(gomock.Any(), 0, 1).Return(PagedResponse{TotalCount: 1, Offset: 0, Body: []byte(`{}`)}, errors.New("request err"))

	d := &Device42DeviceFetcher{
		Limit:       1,
		PageFetcher: mockPageFetcher,
	}

	_, err := d.FetchDevices(context.Background())
	assert.NotNil(t, err)
}
