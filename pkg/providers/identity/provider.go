/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package identity

import (
	"context"

	"github.com/oracle/karpenter-provider-oci/pkg/cache"
	"github.com/oracle/karpenter-provider-oci/pkg/oci"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	ociidentity "github.com/oracle/oci-go-sdk/v65/identity"
)

type Provider interface {
	GetAdMap() map[string]string
	ResolveCompartment(ctx context.Context, compartmentOcid string) (*ociidentity.Compartment, error)
	GetLogicalAdPrefix() string
}

type DefaultProvider struct {
	identityClient       oci.IdentityClient
	adMap                map[string]string
	clusterCompartmentId string
	logicalAdPrefix      string
	compartmentCache     *cache.GetOrLoadCache[*ociidentity.Compartment]
}

func NewProvider(ctx context.Context, clusterCompartmentId string,
	identityClient oci.IdentityClient) (*DefaultProvider, error) {
	p := &DefaultProvider{
		identityClient:       identityClient,
		clusterCompartmentId: clusterCompartmentId,
		compartmentCache:     cache.NewDefaultGetOrLoadCache[*ociidentity.Compartment](),
		adMap:                make(map[string]string),
	}

	err := p.fetchADList(ctx)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *DefaultProvider) GetAdMap() map[string]string {
	return p.adMap
}

func (p *DefaultProvider) fetchADList(ctx context.Context) error {
	if len(p.adMap) > 0 {
		return nil
	}

	// Availability Domains can be listed with a non-root compartment as per tests, so we removed the tenancy finding logic
	resp, err := p.identityClient.ListAvailabilityDomains(ctx, ociidentity.ListAvailabilityDomainsRequest{
		CompartmentId: &p.clusterCompartmentId,
	})

	if err != nil {
		return err
	}

	for _, ad := range resp.Items {
		p.adMap[*ad.Name] = *ad.Id
		p.logicalAdPrefix = utils.ExtractLogicalAdPrefix(*ad.Name)
	}

	return nil
}

func (p *DefaultProvider) ResolveCompartment(ctx context.Context,
	compartmentOcid string) (*ociidentity.Compartment, error) {
	// TODO: life cycle state check
	return p.compartmentCache.GetOrLoad(ctx, compartmentOcid,
		func(c context.Context, key string) (*ociidentity.Compartment, error) {
			resp, err := p.identityClient.GetCompartment(ctx, ociidentity.GetCompartmentRequest{
				CompartmentId: &key,
			})

			if err != nil {
				return nil, err
			}

			return &resp.Compartment, nil
		})
}

func (p *DefaultProvider) GetLogicalAdPrefix() string {
	return p.logicalAdPrefix
}
