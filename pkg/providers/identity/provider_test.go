/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package identity

import (
	"context"
	"errors"
	"testing"

	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	ociidentity "github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/stretchr/testify/assert"
)

const clusterCompId = "ocid1.compartment.oc1..cluster123"

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name        string
		setupFake   func(*fakes.FakeIdentity)
		expectError bool
		validate    func(*testing.T, *DefaultProvider, *fakes.FakeIdentity)
	}{
		{
			name: "happy path",
			setupFake: func(f *fakes.FakeIdentity) {
				// Default setup in NewFakeIdentity provides compartment hierarchy
				// and AD list - should work without additional setup
			},
			expectError: false,
			validate: func(t *testing.T, p *DefaultProvider, f *fakes.FakeIdentity) {
				assert.NotNil(t, p)
				assert.Equal(t, 0, f.GetCompartmentCount.Get()) // cluster -> parent -> tenancy traversal
				assert.Equal(t, 1, f.ListAvailabilityDomainsCount.Get())
				assert.Equal(t, "PHX", p.logicalAdPrefix)
				expectedAdMap := map[string]string{
					"PHX:AD-1": "ocid1.availabilitydomain.oc1..ad1",
					"PHX:AD-2": "ocid1.availabilitydomain.oc1..ad2",
				}
				assert.Equal(t, expectedAdMap, p.adMap)
			},
		},
		{
			name: "list availability domains error",
			setupFake: func(f *fakes.FakeIdentity) {
				f.OnListAvailabilityDomains = func(ctx context.Context,
					req ociidentity.ListAvailabilityDomainsRequest) (ociidentity.ListAvailabilityDomainsResponse, error) {
					return ociidentity.ListAvailabilityDomainsResponse{}, errors.New("AD API error")
				}
			},
			expectError: true,
			validate: func(t *testing.T, p *DefaultProvider, f *fakes.FakeIdentity) {
				assert.Nil(t, p)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fakes.NewFakeIdentity()
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			ctx := context.Background()

			provider, err := NewProvider(ctx, clusterCompId, fakeClient)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validate(t, provider, fakeClient)
		})
	}
}

func TestFetchADList_CacheHit(t *testing.T) {
	fakeClient := fakes.NewFakeIdentity()
	ctx := context.Background()

	provider, err := NewProvider(ctx, clusterCompId, fakeClient)
	assert.NoError(t, err)
	assert.Equal(t, 1, fakeClient.ListAvailabilityDomainsCount.Get())

	// Second call should hit cache (no additional API calls)
	err = provider.fetchADList(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, fakeClient.ListAvailabilityDomainsCount.Get()) // No additional call
}

func TestResolveCompartment(t *testing.T) {
	tests := []struct {
		name            string
		compartmentOcid string
		setupFake       func(*fakes.FakeIdentity)
		expectError     bool
		validate        func(*testing.T, *ociidentity.Compartment, *fakes.FakeIdentity)
	}{
		{
			name:            "resolve existing compartment",
			compartmentOcid: "ocid1.compartment.oc1..parent123",
			setupFake: func(f *fakes.FakeIdentity) {
				// Default setup provides the compartment
			},
			expectError: false,
			validate: func(t *testing.T, comp *ociidentity.Compartment, f *fakes.FakeIdentity) {
				assert.NotNil(t, comp)
				assert.Equal(t, "ocid1.compartment.oc1..parent123", *comp.Id)
				assert.Equal(t, "Parent compartment", *comp.Description)
				assert.Equal(t, 1, f.GetCompartmentCount.Get())
			},
		},
		{
			name:            "compartment not found",
			compartmentOcid: "ocid1.compartment.oc1..nonexistent",
			setupFake: func(f *fakes.FakeIdentity) {
				// Default setup doesn't include this compartment
			},
			expectError: true,
			validate: func(t *testing.T, comp *ociidentity.Compartment, f *fakes.FakeIdentity) {
				assert.Nil(t, comp)
				assert.Equal(t, 1, f.GetCompartmentCount.Get())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fakes.NewFakeIdentity()
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			ctx := context.Background()

			provider, err := NewProvider(ctx, clusterCompId, fakeClient)
			assert.NoError(t, err)

			// Reset counters to test individual ResolveCompartment calls
			fakeClient.GetCompartmentCount = fakes.Counter{}

			result, err := provider.ResolveCompartment(ctx, tt.compartmentOcid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validate(t, result, fakeClient)
		})
	}
}

func TestResolveCompartment_Cache(t *testing.T) {
	fakeClient := fakes.NewFakeIdentity()
	ctx := context.Background()

	provider, err := NewProvider(ctx, clusterCompId, fakeClient)
	assert.NoError(t, err)

	// Reset counters to test ResolveCompartment calls
	fakeClient.GetCompartmentCount = fakes.Counter{}

	// First call should make API request
	result1, err := provider.ResolveCompartment(ctx, "ocid1.compartment.oc1..parent123")
	assert.NoError(t, err)
	assert.NotNil(t, result1)
	assert.Equal(t, 1, fakeClient.GetCompartmentCount.Get())

	// Second call should use cache (no additional API calls)
	result2, err := provider.ResolveCompartment(ctx, "ocid1.compartment.oc1..parent123")
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Equal(t, 1, fakeClient.GetCompartmentCount.Get()) // No additional call

	// Results should be the same
	assert.Equal(t, result1, result2)
}

func TestResolveCompartment_Error(t *testing.T) {
	fakeClient := fakes.NewFakeIdentity()
	ctx := context.Background()

	provider, err := NewProvider(ctx, clusterCompId, fakeClient)
	assert.NoError(t, err)

	// Setup error for ResolveCompartment call
	fakeClient.OnGetCompartment = func(ctx context.Context,
		req ociidentity.GetCompartmentRequest) (ociidentity.GetCompartmentResponse, error) {
		return ociidentity.GetCompartmentResponse{}, errors.New("API error")
	}

	_, err = provider.ResolveCompartment(ctx, "ocid1.compartment.oc1..parent123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestGetAdMap(t *testing.T) {
	fakeClient := fakes.NewFakeIdentity()
	ctx := context.Background()

	provider, err := NewProvider(ctx, clusterCompId, fakeClient)
	assert.NoError(t, err)

	adMap := provider.GetAdMap()
	expectedAdMap := map[string]string{
		"PHX:AD-1": "ocid1.availabilitydomain.oc1..ad1",
		"PHX:AD-2": "ocid1.availabilitydomain.oc1..ad2",
	}
	assert.Equal(t, expectedAdMap, adMap)
}

func TestGetLogicalAdPrefix(t *testing.T) {
	fakeClient := fakes.NewFakeIdentity()
	ctx := context.Background()

	provider, err := NewProvider(ctx, clusterCompId, fakeClient)
	assert.NoError(t, err)

	prefix := provider.GetLogicalAdPrefix()
	assert.Equal(t, "PHX", prefix)
}
