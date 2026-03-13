/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package fakes

import (
	"context"

	ocikms "github.com/oracle/oci-go-sdk/v65/keymanagement"
)

// FakeKms implements oci.KmsClient for tests
type FakeKms struct {
	getResp  ocikms.GetKeyResponse
	getErr   error
	GetCount Counter
	OnGet    func(context.Context, ocikms.GetKeyRequest) (ocikms.GetKeyResponse, error)
}

func (f *FakeKms) GetKey(ctx context.Context, req ocikms.GetKeyRequest) (ocikms.GetKeyResponse, error) {
	f.GetCount.Inc()
	if f.OnGet != nil {
		return f.OnGet(ctx, req)
	}
	if f.getErr != nil {
		return ocikms.GetKeyResponse{}, f.getErr
	}
	return f.getResp, nil
}
