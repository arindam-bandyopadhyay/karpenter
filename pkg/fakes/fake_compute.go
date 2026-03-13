/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package fakes

import (
	"context"

	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	ociwr "github.com/oracle/oci-go-sdk/v65/workrequests"
	"github.com/samber/lo"
)

// FakeCompute implements oci.ComputeClient for tests.
// nolint:lll
type FakeCompute struct {
	LastLaunchReq                          ocicore.LaunchInstanceRequest
	LaunchResp                             ocicore.LaunchInstanceResponse
	LaunchErr                              error
	getResp                                ocicore.GetInstanceResponse
	GetErr                                 error
	listInstancesResp                      ocicore.ListInstancesResponse
	ListInstancesErr                       error
	ListResp                               ocicore.ListComputeClustersResponse
	GetImageResp                           ocicore.GetImageResponse
	GetImageErr                            error
	ListImagesResp                         ocicore.ListImagesResponse
	ListImagesErr                          error
	ListImageShapeCompatibilityEntriesResp ocicore.ListImageShapeCompatibilityEntriesResponse
	ListImageShapeCompatibilityEntriesErr  error
	VnicPages                              [][]ocicore.VnicAttachment
	BootPages                              [][]ocicore.BootVolumeAttachment
	CapacityResPages                       [][]ocicore.ComputeCapacityReservationSummary
	listShapesResp                         ocicore.ListShapesResponse
	ListShapesErr                          error
	TerminateErr                           error

	LaunchCount                             Counter
	GetCount                                Counter
	ListInstancesCount                      Counter
	ListShapesCount                         Counter
	ListVnicCount                           Counter
	ListBootCount                           Counter
	TerminateCount                          Counter
	GetClusterCount                         Counter
	GetComputeClusterCount                  Counter
	ListClustersCount                       Counter
	ListComputeClustersCount                Counter
	GetImageCount                           Counter
	ListImagesCount                         Counter
	ListImageShapeCompatibilityEntriesCount Counter

	OnLaunch                             func(context.Context, ocicore.LaunchInstanceRequest) (ocicore.LaunchInstanceResponse, error)
	OnGet                                func(context.Context, ocicore.GetInstanceRequest) (ocicore.GetInstanceResponse, error)
	OnListInstances                      func(context.Context, ocicore.ListInstancesRequest) (ocicore.ListInstancesResponse, error)
	OnListShapes                         func(context.Context, ocicore.ListShapesRequest) (ocicore.ListShapesResponse, error)
	OnListVnics                          func(context.Context, ocicore.ListVnicAttachmentsRequest) (ocicore.ListVnicAttachmentsResponse, error)
	OnListBoot                           func(context.Context, ocicore.ListBootVolumeAttachmentsRequest) (ocicore.ListBootVolumeAttachmentsResponse, error)
	OnTerminate                          func(context.Context, ocicore.TerminateInstanceRequest) (ocicore.TerminateInstanceResponse, error)
	OnGetImage                           func(context.Context, ocicore.GetImageRequest) (ocicore.GetImageResponse, error)
	OnListImages                         func(context.Context, ocicore.ListImagesRequest) (ocicore.ListImagesResponse, error)
	OnListImageShapeCompatibilityEntries func(context.Context, ocicore.ListImageShapeCompatibilityEntriesRequest) (ocicore.ListImageShapeCompatibilityEntriesResponse, error)
	OnGetComputeCapacityReservation      func(context.Context, ocicore.GetComputeCapacityReservationRequest) (ocicore.GetComputeCapacityReservationResponse, error)
	OnListComputeCapacityReservations    func(context.Context, ocicore.ListComputeCapacityReservationsRequest) (ocicore.ListComputeCapacityReservationsResponse, error)
	OnGetComputeCluster                  func(context.Context, ocicore.GetComputeClusterRequest) (ocicore.GetComputeClusterResponse, error)
	OnListComputeClusters                func(context.Context, ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error)
}

func (f *FakeCompute) LaunchInstance(ctx context.Context,
	req ocicore.LaunchInstanceRequest) (ocicore.LaunchInstanceResponse, error) {
	f.LaunchCount.Inc()
	f.LastLaunchReq = req
	if f.OnLaunch != nil {
		return f.OnLaunch(ctx, req)
	}
	if f.LaunchErr != nil {
		return ocicore.LaunchInstanceResponse{}, f.LaunchErr
	}
	if f.LaunchResp.Instance.Id == nil {
		f.LaunchResp.Instance = ocicore.Instance{Id: lo.ToPtr("ocid1.instance.oc1..new")}
	}
	if f.LaunchResp.Etag == nil {
		f.LaunchResp.Etag = lo.ToPtr("etag-new")
	}
	if f.LaunchResp.OpcWorkRequestId == nil {
		f.LaunchResp.OpcWorkRequestId = lo.ToPtr("wr1")
	}
	return f.LaunchResp, nil
}

func (f *FakeCompute) GetInstance(ctx context.Context, req ocicore.GetInstanceRequest) (
	ocicore.GetInstanceResponse, error) {
	f.GetCount.Inc()
	if f.OnGet != nil {
		return f.OnGet(ctx, req)
	}
	if f.GetErr != nil {
		return ocicore.GetInstanceResponse{}, f.GetErr
	}
	return f.getResp, nil
}

func (f *FakeCompute) ListInstances(ctx context.Context,
	req ocicore.ListInstancesRequest) (ocicore.ListInstancesResponse, error) {
	f.ListInstancesCount.Inc()
	if f.OnListInstances != nil {
		return f.OnListInstances(ctx, req)
	}
	if f.ListInstancesErr != nil {
		return ocicore.ListInstancesResponse{}, f.ListInstancesErr
	}
	return f.listInstancesResp, nil
}

func (f *FakeCompute) ListShapes(ctx context.Context,
	req ocicore.ListShapesRequest) (ocicore.ListShapesResponse, error) {
	f.ListShapesCount.Inc()
	if f.OnListShapes != nil {
		return f.OnListShapes(ctx, req)
	}
	if f.ListShapesErr != nil {
		return ocicore.ListShapesResponse{}, f.ListShapesErr
	}
	return f.listShapesResp, nil
}

func (f *FakeCompute) ListVnicAttachments(ctx context.Context,
	req ocicore.ListVnicAttachmentsRequest) (ocicore.ListVnicAttachmentsResponse, error) {
	f.ListVnicCount.Inc()
	if f.OnListVnics != nil {
		return f.OnListVnics(ctx, req)
	}
	index := pageIndex(req.Page)
	resp := ocicore.ListVnicAttachmentsResponse{}
	if index < len(f.VnicPages) {
		resp.Items = f.VnicPages[index]
		if index+1 < len(f.VnicPages) {
			resp.OpcNextPage = lo.ToPtr(nextPageToken(index + 1))
		}
	}
	return resp, nil
}

func (f *FakeCompute) ListBootVolumeAttachments(ctx context.Context,
	req ocicore.ListBootVolumeAttachmentsRequest) (ocicore.ListBootVolumeAttachmentsResponse, error) {
	f.ListBootCount.Inc()
	if f.OnListBoot != nil {
		return f.OnListBoot(ctx, req)
	}
	index := pageIndex(req.Page)
	resp := ocicore.ListBootVolumeAttachmentsResponse{}
	if index < len(f.BootPages) {
		resp.Items = f.BootPages[index]
		if index+1 < len(f.BootPages) {
			resp.OpcNextPage = lo.ToPtr(nextPageToken(index + 1))
		}
	}
	return resp, nil
}

func (f *FakeCompute) TerminateInstance(ctx context.Context, req ocicore.TerminateInstanceRequest) (
	ocicore.TerminateInstanceResponse, error) {
	f.TerminateCount.Inc()
	if f.OnTerminate != nil {
		return f.OnTerminate(ctx, req)
	}
	if f.TerminateErr != nil {
		return ocicore.TerminateInstanceResponse{}, f.TerminateErr
	}
	return ocicore.TerminateInstanceResponse{}, nil
}
func (f *FakeCompute) GetComputeCapacityReservation(
	ctx context.Context,
	request ocicore.GetComputeCapacityReservationRequest,
) (ocicore.GetComputeCapacityReservationResponse, error) {
	if f.OnGetComputeCapacityReservation != nil {
		return f.OnGetComputeCapacityReservation(ctx, request)
	}
	return ocicore.GetComputeCapacityReservationResponse{}, nil
}

func (f *FakeCompute) ListComputeCapacityReservations(ctx context.Context,
	req ocicore.ListComputeCapacityReservationsRequest) (ocicore.ListComputeCapacityReservationsResponse, error) {
	if f.OnListComputeCapacityReservations != nil {
		return f.OnListComputeCapacityReservations(ctx, req)
	}
	index := pageIndex(req.Page)
	resp := ocicore.ListComputeCapacityReservationsResponse{}
	if index < len(f.CapacityResPages) {
		resp.Items = f.CapacityResPages[index]
		if index+1 < len(f.CapacityResPages) {
			resp.OpcNextPage = lo.ToPtr(nextPageToken(index + 1))
		}
	}
	return resp, nil
}

func (f *FakeCompute) GetComputeCluster(ctx context.Context, req ocicore.GetComputeClusterRequest) (
	ocicore.GetComputeClusterResponse, error) {
	f.GetClusterCount.Inc()
	f.GetComputeClusterCount.Inc()
	if f.OnGetComputeCluster != nil {
		return f.OnGetComputeCluster(ctx, req)
	}
	return ocicore.GetComputeClusterResponse{}, nil
}

func (f *FakeCompute) ListComputeClusters(ctx context.Context, req ocicore.ListComputeClustersRequest) (
	ocicore.ListComputeClustersResponse, error) {
	f.ListClustersCount.Inc()
	f.ListComputeClustersCount.Inc()
	if f.OnListComputeClusters != nil {
		return f.OnListComputeClusters(ctx, req)
	}
	resp := ocicore.ListComputeClustersResponse{
		ComputeClusterCollection: ocicore.ComputeClusterCollection{
			Items: []ocicore.ComputeClusterSummary{},
		},
	}
	if len(f.ListResp.ComputeClusterCollection.Items) > 0 {
		items := f.ListResp.ComputeClusterCollection.Items
		if req.DisplayName != nil {
			filteredItems := make([]ocicore.ComputeClusterSummary, 0)
			for _, item := range items {
				if item.DisplayName != nil && *item.DisplayName == *req.DisplayName {
					filteredItems = append(filteredItems, item)
				}
			}
			items = filteredItems
		}
		resp.ComputeClusterCollection.Items = items
		resp.OpcNextPage = f.ListResp.OpcNextPage
	}
	return resp, nil
}

func (f *FakeCompute) GetImage(ctx context.Context, req ocicore.GetImageRequest) (ocicore.GetImageResponse, error) {
	f.GetImageCount.Inc()
	if f.OnGetImage != nil {
		return f.OnGetImage(ctx, req)
	}
	if f.GetImageErr != nil {
		return ocicore.GetImageResponse{}, f.GetImageErr
	}
	return f.GetImageResp, nil
}

func (f *FakeCompute) ListImages(ctx context.Context, req ocicore.ListImagesRequest) (
	ocicore.ListImagesResponse, error) {
	f.ListImagesCount.Inc()
	if f.OnListImages != nil {
		return f.OnListImages(ctx, req)
	}
	if f.ListImagesErr != nil {
		return ocicore.ListImagesResponse{}, f.ListImagesErr
	}
	resp := f.ListImagesResp
	if req.OperatingSystem != nil || req.OperatingSystemVersion != nil {
		filteredItems := make([]ocicore.Image, 0)
		for _, item := range resp.Items {
			if req.OperatingSystem != nil && item.OperatingSystem != nil && *req.OperatingSystem != *item.OperatingSystem {
				continue
			}
			if req.OperatingSystemVersion != nil && item.OperatingSystemVersion != nil &&
				*req.OperatingSystemVersion != *item.OperatingSystemVersion {
				continue
			}
			filteredItems = append(filteredItems, item)
		}
		resp.Items = filteredItems
	}
	return resp, nil
}

func (f *FakeCompute) ListImageShapeCompatibilityEntries(ctx context.Context,
	req ocicore.ListImageShapeCompatibilityEntriesRequest) (ocicore.ListImageShapeCompatibilityEntriesResponse, error) {
	f.ListImageShapeCompatibilityEntriesCount.Inc()
	if f.OnListImageShapeCompatibilityEntries != nil {
		return f.OnListImageShapeCompatibilityEntries(ctx, req)
	}
	if f.ListImageShapeCompatibilityEntriesErr != nil {
		return ocicore.ListImageShapeCompatibilityEntriesResponse{}, f.ListImageShapeCompatibilityEntriesErr
	}
	return f.ListImageShapeCompatibilityEntriesResp, nil
}

// FakeWorkRequest supports:
// - Recording the last GetWorkRequest request (for end-to-end request assertions)
// - Canned responses and error injection for GetWorkRequest
// - Optional hook functions for flexible behavior
// - Call counters for verifying polling behavior
type FakeWorkRequest struct {
	GetResp        ociwr.GetWorkRequestResponse
	GetErr         error
	ListErrorsResp ociwr.ListWorkRequestErrorsResponse
	ListErrorsErr  error
	GetCount       Counter
	ListCount      Counter
	OnGet          func(context.Context, ociwr.GetWorkRequestRequest) (ociwr.GetWorkRequestResponse, error)
	OnList         func(context.Context, ociwr.ListWorkRequestErrorsRequest) (ociwr.ListWorkRequestErrorsResponse, error)
}

func (f *FakeWorkRequest) GetWorkRequest(ctx context.Context, req ociwr.GetWorkRequestRequest) (
	ociwr.GetWorkRequestResponse, error) {
	f.GetCount.Inc()
	if f.OnGet != nil {
		return f.OnGet(ctx, req)
	}
	if f.GetErr != nil {
		return ociwr.GetWorkRequestResponse{}, f.GetErr
	}
	return f.GetResp, nil
}

func (f *FakeWorkRequest) ListWorkRequestErrors(ctx context.Context, req ociwr.ListWorkRequestErrorsRequest) (
	ociwr.ListWorkRequestErrorsResponse, error) {
	f.ListCount.Inc()
	if f.OnList != nil {
		return f.OnList(ctx, req)
	}
	if f.ListErrorsErr != nil {
		return ociwr.ListWorkRequestErrorsResponse{}, f.ListErrorsErr
	}
	return f.ListErrorsResp, nil
}
