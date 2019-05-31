package device42

import (
	"context"
	"fmt"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFetchDevices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPaginator := NewMockPaginator(ctrl)
	mockPaginator.EXPECT().BatchPagedRequests(gomock.Any()).Return([][]byte{[]byte(`{"ips": [{"ip": "192.168.1.1", "device_id": 1, "subnet_id": 1}, {"ip": "192.168.1.2", "device_id": 2, "subnet_id": 2}]}`)}, nil)
	d := &Device42DeviceFetcher{
		Paginator: mockPaginator,
	}

	assets, err := d.FetchDevices(context.Background())
	assert.ElementsMatch(t, []domain.Device{domain.Device{IP: "192.168.1.1", ID: "1", SubnetID: "1"}, domain.Device{IP: "192.168.1.2", ID: "2", SubnetID: "2"}}, assets)
	assert.Nil(t, err)
}

func TestFetchDevicesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPaginator := NewMockPaginator(ctrl)
	mockPaginator.EXPECT().BatchPagedRequests(gomock.Any()).Return(nil, fmt.Errorf("batch error"))
	d := &Device42DeviceFetcher{
		Paginator: mockPaginator,
	}

	_, err := d.FetchDevices(context.Background())
	assert.NotNil(t, err)
}

func TestFetchDevicesUnmarshalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPaginator := NewMockPaginator(ctrl)
	mockPaginator.EXPECT().BatchPagedRequests(gomock.Any()).Return([][]byte{[]byte("notanip")}, nil)
	d := &Device42DeviceFetcher{
		Paginator: mockPaginator,
	}

	_, err := d.FetchDevices(context.Background())
	assert.NotNil(t, err)
}
