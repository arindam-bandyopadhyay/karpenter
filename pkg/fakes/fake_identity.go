/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package fakes

import (
	"context"
	"errors"
	"net/http"

	ociidentity "github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/samber/lo"
)

// FakeIdentity implements oci.IdentityClient for tests.
// Supports canned compartment hierarchy and availability domains.
// nolint:lll
type FakeIdentity struct {
	// Compartment hierarchy for tenancy traversal
	compartments map[string]ociidentity.Compartment

	// Availability domains response
	listADResp ociidentity.ListAvailabilityDomainsResponse
	listADErr  error

	// Counters
	GetCompartmentCount          Counter
	ListAvailabilityDomainsCount Counter

	// Optional hook funcs (if set, they override canned behavior)
	OnGetCompartment          func(context.Context, ociidentity.GetCompartmentRequest) (ociidentity.GetCompartmentResponse, error)
	OnListAvailabilityDomains func(context.Context, ociidentity.ListAvailabilityDomainsRequest) (ociidentity.ListAvailabilityDomainsResponse, error)
}

// NewFakeIdentity creates a fake identity client with canned compartment hierarchy
// that simulates tenancy traversal (child -> parent -> tenancy root).
func NewFakeIdentity() *FakeIdentity {
	f := &FakeIdentity{
		compartments: make(map[string]ociidentity.Compartment),
	}

	// Setup canned compartment hierarchy:
	// tenancy (root) -> parent-comp -> cluster-comp
	tenancyId := "ocid1.tenancy.oc1..tenancy123"
	parentCompId := "ocid1.compartment.oc1..parent123"
	clusterCompId := "ocid1.compartment.oc1..cluster123"

	f.compartments[tenancyId] = ociidentity.Compartment{
		Id:          lo.ToPtr(tenancyId),
		Description: lo.ToPtr("Root tenancy"),
		// Root has no CompartmentId (nil)
	}

	f.compartments[parentCompId] = ociidentity.Compartment{
		Id:            lo.ToPtr(parentCompId),
		Description:   lo.ToPtr("Parent compartment"),
		CompartmentId: lo.ToPtr(tenancyId), // Parent points to tenancy
	}

	f.compartments[clusterCompId] = ociidentity.Compartment{
		Id:            lo.ToPtr(clusterCompId),
		Description:   lo.ToPtr("Cluster compartment"),
		CompartmentId: lo.ToPtr(parentCompId), // Child points to parent
	}

	// Setup availability domains
	f.listADResp = ociidentity.ListAvailabilityDomainsResponse{
		RawResponse: &http.Response{
			StatusCode: 200,
		},
		Items: []ociidentity.AvailabilityDomain{
			{
				Id:   lo.ToPtr("ocid1.availabilitydomain.oc1..ad1"),
				Name: lo.ToPtr("PHX:AD-1"),
			},
			{
				Id:   lo.ToPtr("ocid1.availabilitydomain.oc1..ad2"),
				Name: lo.ToPtr("PHX:AD-2"),
			},
		},
	}

	return f
}

func (f *FakeIdentity) GetCompartment(ctx context.Context,
	request ociidentity.GetCompartmentRequest) (ociidentity.GetCompartmentResponse, error) {
	f.GetCompartmentCount.Inc()
	if f.OnGetCompartment != nil {
		return f.OnGetCompartment(ctx, request)
	}

	comp, exists := f.compartments[*request.CompartmentId]
	if !exists {
		return ociidentity.GetCompartmentResponse{}, errors.New("compartment not found")
	}

	return ociidentity.GetCompartmentResponse{
		Compartment: comp,
	}, nil
}

func (f *FakeIdentity) ListAvailabilityDomains(ctx context.Context,
	request ociidentity.ListAvailabilityDomainsRequest) (ociidentity.ListAvailabilityDomainsResponse, error) {
	f.ListAvailabilityDomainsCount.Inc()
	if f.OnListAvailabilityDomains != nil {
		return f.OnListAvailabilityDomains(ctx, request)
	}

	if f.listADErr != nil {
		return ociidentity.ListAvailabilityDomainsResponse{}, f.listADErr
	}

	return f.listADResp, nil
}
