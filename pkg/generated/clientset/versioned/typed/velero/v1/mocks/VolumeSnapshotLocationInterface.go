// Code generated by mockery v2.30.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	types "k8s.io/apimachinery/pkg/types"

	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"

	watch "k8s.io/apimachinery/pkg/watch"
)

// VolumeSnapshotLocationInterface is an autogenerated mock type for the VolumeSnapshotLocationInterface type
type VolumeSnapshotLocationInterface struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, volumeSnapshotLocation, opts
func (_m *VolumeSnapshotLocationInterface) Create(ctx context.Context, volumeSnapshotLocation *v1.VolumeSnapshotLocation, opts metav1.CreateOptions) (*v1.VolumeSnapshotLocation, error) {
	ret := _m.Called(ctx, volumeSnapshotLocation, opts)

	var r0 *v1.VolumeSnapshotLocation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.CreateOptions) (*v1.VolumeSnapshotLocation, error)); ok {
		return rf(ctx, volumeSnapshotLocation, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.CreateOptions) *v1.VolumeSnapshotLocation); ok {
		r0 = rf(ctx, volumeSnapshotLocation, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.VolumeSnapshotLocation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.CreateOptions) error); ok {
		r1 = rf(ctx, volumeSnapshotLocation, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, name, opts
func (_m *VolumeSnapshotLocationInterface) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	ret := _m.Called(ctx, name, opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.DeleteOptions) error); ok {
		r0 = rf(ctx, name, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteCollection provides a mock function with given fields: ctx, opts, listOpts
func (_m *VolumeSnapshotLocationInterface) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	ret := _m.Called(ctx, opts, listOpts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.DeleteOptions, metav1.ListOptions) error); ok {
		r0 = rf(ctx, opts, listOpts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, name, opts
func (_m *VolumeSnapshotLocationInterface) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.VolumeSnapshotLocation, error) {
	ret := _m.Called(ctx, name, opts)

	var r0 *v1.VolumeSnapshotLocation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.GetOptions) (*v1.VolumeSnapshotLocation, error)); ok {
		return rf(ctx, name, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.GetOptions) *v1.VolumeSnapshotLocation); ok {
		r0 = rf(ctx, name, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.VolumeSnapshotLocation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, metav1.GetOptions) error); ok {
		r1 = rf(ctx, name, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, opts
func (_m *VolumeSnapshotLocationInterface) List(ctx context.Context, opts metav1.ListOptions) (*v1.VolumeSnapshotLocationList, error) {
	ret := _m.Called(ctx, opts)

	var r0 *v1.VolumeSnapshotLocationList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) (*v1.VolumeSnapshotLocationList, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) *v1.VolumeSnapshotLocationList); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.VolumeSnapshotLocationList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Patch provides a mock function with given fields: ctx, name, pt, data, opts, subresources
func (_m *VolumeSnapshotLocationInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*v1.VolumeSnapshotLocation, error) {
	_va := make([]interface{}, len(subresources))
	for _i := range subresources {
		_va[_i] = subresources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, name, pt, data, opts)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *v1.VolumeSnapshotLocation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) (*v1.VolumeSnapshotLocation, error)); ok {
		return rf(ctx, name, pt, data, opts, subresources...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) *v1.VolumeSnapshotLocation); ok {
		r0 = rf(ctx, name, pt, data, opts, subresources...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.VolumeSnapshotLocation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) error); ok {
		r1 = rf(ctx, name, pt, data, opts, subresources...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, volumeSnapshotLocation, opts
func (_m *VolumeSnapshotLocationInterface) Update(ctx context.Context, volumeSnapshotLocation *v1.VolumeSnapshotLocation, opts metav1.UpdateOptions) (*v1.VolumeSnapshotLocation, error) {
	ret := _m.Called(ctx, volumeSnapshotLocation, opts)

	var r0 *v1.VolumeSnapshotLocation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.UpdateOptions) (*v1.VolumeSnapshotLocation, error)); ok {
		return rf(ctx, volumeSnapshotLocation, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.UpdateOptions) *v1.VolumeSnapshotLocation); ok {
		r0 = rf(ctx, volumeSnapshotLocation, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.VolumeSnapshotLocation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.UpdateOptions) error); ok {
		r1 = rf(ctx, volumeSnapshotLocation, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStatus provides a mock function with given fields: ctx, volumeSnapshotLocation, opts
func (_m *VolumeSnapshotLocationInterface) UpdateStatus(ctx context.Context, volumeSnapshotLocation *v1.VolumeSnapshotLocation, opts metav1.UpdateOptions) (*v1.VolumeSnapshotLocation, error) {
	ret := _m.Called(ctx, volumeSnapshotLocation, opts)

	var r0 *v1.VolumeSnapshotLocation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.UpdateOptions) (*v1.VolumeSnapshotLocation, error)); ok {
		return rf(ctx, volumeSnapshotLocation, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.UpdateOptions) *v1.VolumeSnapshotLocation); ok {
		r0 = rf(ctx, volumeSnapshotLocation, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.VolumeSnapshotLocation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.VolumeSnapshotLocation, metav1.UpdateOptions) error); ok {
		r1 = rf(ctx, volumeSnapshotLocation, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Watch provides a mock function with given fields: ctx, opts
func (_m *VolumeSnapshotLocationInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	ret := _m.Called(ctx, opts)

	var r0 watch.Interface
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) (watch.Interface, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) watch.Interface); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(watch.Interface)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewVolumeSnapshotLocationInterface creates a new instance of VolumeSnapshotLocationInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewVolumeSnapshotLocationInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *VolumeSnapshotLocationInterface {
	mock := &VolumeSnapshotLocationInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
