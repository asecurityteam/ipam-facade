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
	mockIterator := NewMockIterator(ctrl)
	mockIterator.EXPECT().Next().Return(true)
	mockIterator.EXPECT().Current().Return(PagedResponse{Body: []byte(`{"ips": [{"ip": "192.168.1.1", "device_id": 1, "subnet_id": 1}, {"ip": "192.168.1.2", "device_id": 2, "subnet_id": 2}]}`)})
	mockIterator.EXPECT().Next().Return(false)
	mockIterator.EXPECT().Close().Return(nil)

	d := &Device42DeviceFetcher{
		Iterator: mockIterator,
	}

	assets, err := d.FetchDevices(context.Background())
	assert.ElementsMatch(t, []domain.Device{domain.Device{IP: "192.168.1.1", ID: "1", SubnetID: "1"}, domain.Device{IP: "192.168.1.2", ID: "2", SubnetID: "2"}}, assets)
	assert.Nil(t, err)
}
func TestFetchDevicesMultiple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIterator := NewMockIterator(ctrl)
	mockIterator.EXPECT().Next().Return(true)
	mockIterator.EXPECT().Current().Return(PagedResponse{Body: []byte(`{"ips": [{"ip": "192.168.1.1", "device_id": 1, "subnet_id": 1}, {"ip": "192.168.1.2", "device_id": 2, "subnet_id": 2}]}`)})
	mockIterator.EXPECT().Next().Return(true)
	mockIterator.EXPECT().Current().Return(PagedResponse{Body: []byte(`{"ips": [{"ip": "192.168.1.3", "device_id": 3, "subnet_id": 3}]}`)})
	mockIterator.EXPECT().Next().Return(false)
	mockIterator.EXPECT().Close().Return(nil)

	d := &Device42DeviceFetcher{
		Iterator: mockIterator,
	}

	assets, err := d.FetchDevices(context.Background())
	assert.ElementsMatch(t, []domain.Device{domain.Device{IP: "192.168.1.1", ID: "1", SubnetID: "1"}, domain.Device{IP: "192.168.1.2", ID: "2", SubnetID: "2"}, domain.Device{IP: "192.168.1.3", ID: "3", SubnetID: "3"}}, assets)
	assert.Nil(t, err)
}
func TestFetchDevicesUnmarshalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIterator := NewMockIterator(ctrl)
	mockIterator.EXPECT().Next().Return(true)
	mockIterator.EXPECT().Current().Return(PagedResponse{Body: []byte("notanip")})

	d := &Device42DeviceFetcher{
		Iterator: mockIterator,
	}

	_, err := d.FetchDevices(context.Background())
	assert.NotNil(t, err)
}

func TestFetchDevicesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIterator := NewMockIterator(ctrl)
	mockIterator.EXPECT().Next().Return(false)
	mockIterator.EXPECT().Close().Return(errors.New("iterator error"))

	d := &Device42DeviceFetcher{
		Iterator: mockIterator,
	}

	_, err := d.FetchDevices(context.Background())
	assert.NotNil(t, err)
}
