/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package kms

import (
	"context"
	"fmt"
	"sync"
	"testing"

	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/karpenter-provider-oci/pkg/oci"
	"github.com/oracle/oci-go-sdk/v65/common"
	ocikms "github.com/oracle/oci-go-sdk/v65/keymanagement"
)

const testValidKeyOCID = "ocid1.key.oc1.iad.vprefix.ext1.ext2"

func newTestProvider(t *testing.T) *DefaultProvider {
	t.Helper()
	cfg := common.NewRawConfigurationProvider(
		"ocid1.tenancy.oc1..dummy", // tenancy
		"ocid1.user.oc1..dummy",    // user
		"us-phoenix-1",             // region (string form, not region code)
		"d3:ad:be:ef",              // fingerprint
		"-----BEGIN PRIVATE KEY-----\nMIICfake\n-----END PRIVATE KEY-----", // dummy key; not used in tests
		nil,
	)
	p, err := NewProvider(context.TODO(), "ocid1.compartment.oc1..dummy", cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	return p
}

func TestProvider_ResolveKmsKeyConfig_Table(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newTestProvider(t)

	keyCached := "ocid1.key.oc1.iad.vaultA.someext.more"
	keyInvalid := "ocid1.key.oc1" // invalid, less than 6 parts

	// Seed via cache.GetOrLoad to avoid touching unexported fields
	display := "my-key"
	_, seedErr := p.kmsOcidCache.GetOrLoad(ctx, keyCached, func(context.Context, string) (*ocikms.Key, error) {
		return &ocikms.Key{Id: &keyCached, DisplayName: &display}, nil
	})
	if seedErr != nil {
		t.Fatalf("seed cache error = %v", seedErr)
	}

	tests := []struct {
		name     string
		input    *ociv1beta1.KmsKeyConfig
		wantNil  bool
		wantErr  bool
		wantOcid string
		wantName string
		cacheHit bool
	}{
		{
			name:    "nil input returns nil result and no error",
			input:   nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name:     "cached key returns resolve result without invoking loader",
			input:    &ociv1beta1.KmsKeyConfig{KmsKeyId: &keyCached},
			wantNil:  false,
			wantErr:  false,
			wantOcid: keyCached,
			wantName: display,
			cacheHit: true,
		},
		{
			name:    "invalid ocid returns error",
			input:   &ociv1beta1.KmsKeyConfig{KmsKeyId: &keyInvalid},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.ResolveKmsKeyConfig(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				if assert.NotNil(t, got) {
					assert.Equal(t, tt.wantOcid, got.Ocid)
					assert.Equal(t, tt.wantName, got.DisplayName)
				}
			}
		})
	}
}

func TestProvider_GetKmsEndpointFromKeyId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ocid       string
		wantErr    bool
		wantPrefix string
		wantRegion string
	}{
		{
			name:       "iad region with vault prefix",
			ocid:       "ocid1.key.oc1.iad.vaultA.ext1.ext2",
			wantErr:    false,
			wantPrefix: "vaultA",
			wantRegion: string(common.StringToRegion("iad")),
		},
		{
			name:       "phx region with vault prefix",
			ocid:       "ocid1.key.oc1.phx.VAULTZ.ext1.ext2",
			wantErr:    false,
			wantPrefix: "VAULTZ",
			wantRegion: string(common.StringToRegion("phx")),
		},
		{
			name:    "invalid ocid too few parts",
			ocid:    "ocid1.key.oc1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := getKmsEndpointFromKeyId(tt.ocid)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			// Expect: https://<vaultPrefix>-management.kms.<region>.oraclecloud.com
			expected := fmt.Sprintf("https://%s-management.kms.%s.oraclecloud.com", tt.wantPrefix, tt.wantRegion)
			assert.Equal(t, expected, endpoint)
		})
	}
}

func TestProvider_GetExtensions(t *testing.T) {
	t.Parallel()

	valid := testValidKeyOCID
	parts, err := getExtensions(valid)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(parts), 6)
	assert.Equal(t, "iad", parts[3])
	assert.Equal(t, "vprefix", parts[4])

	invalid := "ocid1.key.oc1"
	_, err = getExtensions(invalid)
	assert.Error(t, err)
}

func TestProvider_NewProvider_InitializesInternalCaches(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)
	// Basic sanity on internals
	assert.NotNil(t, p.kmsOcidCache)
	assert.NotNil(t, p.kmsFilterCache)
	assert.NotNil(t, p.kmsClientCache)

	// Ensure cache works end-to-end for a synthetic key (no network)
	ctx := context.Background()
	keyID := "ocid1.key.oc1.iad.vA.x.y"
	name := "seeded"
	_, seedErr := p.kmsOcidCache.GetOrLoad(ctx, keyID, func(context.Context, string) (*ocikms.Key, error) {
		return &ocikms.Key{Id: &keyID, DisplayName: &name}, nil
	})
	if seedErr != nil {
		t.Fatalf("seed cache error = %v", seedErr)
	}

	res, err := p.ResolveKmsKeyConfig(ctx, &ociv1beta1.KmsKeyConfig{KmsKeyId: &keyID})
	assert.NoError(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, keyID, res.Ocid)
		assert.Equal(t, name, res.DisplayName)
	}
}

func TestProvider_GetKmsClient_Cache(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)

	ep1, _ := getKmsEndpointFromKeyId("ocid1.key.oc1.iad.vaultX.ext1.ext2")
	initialLen := len(p.kmsClientCache)
	p.kmsClientCache[ep1] = &fakes.FakeKms{}
	c1, err := p.getKmsClient(ep1)
	require.NoError(t, err)
	c2, err := p.getKmsClient(ep1)
	require.NoError(t, err)
	assert.Same(t, c1, c2)

	ep2, _ := getKmsEndpointFromKeyId("ocid1.key.oc1.iad.vaultY.ext1.ext2")
	p.kmsClientCache[ep2] = &fakes.FakeKms{}
	c3, err := p.getKmsClient(ep2)
	require.NoError(t, err)
	assert.NotSame(t, c1, c3)
	assert.Equal(t, initialLen+2, len(p.kmsClientCache))
}

func TestProvider_GetKmsClient_DisabledServiceError(t *testing.T) {
	// This test mutates a global in the OCI SDK, so do not run in parallel.
	p := newTestProvider(t)

	// Make only a different service enabled so "keymanagement" is considered disabled.
	prev := common.OciSdkEnabledServicesMap
	t.Cleanup(func() { common.OciSdkEnabledServicesMap = prev })
	common.OciSdkEnabledServicesMap = map[string]bool{"compute": true}

	endpoint, _ := getKmsEndpointFromKeyId("ocid1.key.oc1.iad.VPREFIX.ext1.ext2")
	_, err := p.getKmsClient(endpoint)
	assert.Error(t, err, "expected error when keymanagement service is disabled by SDK global map")
}

func TestProvider_ResolveKmsKeyConfig_ClientError(t *testing.T) {
	// This test also relies on SDK global. Do not run in parallel.
	p := newTestProvider(t)
	ctx := context.Background()

	// Disable keymanagement to force client construction error inside resolve path.
	prev := common.OciSdkEnabledServicesMap
	t.Cleanup(func() { common.OciSdkEnabledServicesMap = prev })
	common.OciSdkEnabledServicesMap = map[string]bool{"objectstorage": true}

	validKey := "ocid1.key.oc1.iad.vpx.ext1.ext2" // syntactically valid -> endpoint derivable
	res, err := p.ResolveKmsKeyConfig(ctx, &ociv1beta1.KmsKeyConfig{KmsKeyId: &validKey})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestProvider_ResolveKmsKeyConfig_OCIGetKey_Success(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	keyID := testValidKeyOCID
	endpoint, err := getKmsEndpointFromKeyId(keyID)
	assert.NoError(t, err)

	called := 0

	p.kmsClientCache = make(map[string]oci.KmsClient)
	fk := &fakes.FakeKms{OnGet: func(ctx context.Context, r ocikms.GetKeyRequest) (ocikms.GetKeyResponse, error) {
		called++
		display := "from-oci"
		return ocikms.GetKeyResponse{Key: ocikms.Key{Id: &keyID, DisplayName: &display}}, nil
	}}
	p.kmsClientCache[endpoint] = fk

	res, err := p.ResolveKmsKeyConfig(ctx, &ociv1beta1.KmsKeyConfig{KmsKeyId: &keyID})
	assert.NoError(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, keyID, res.Ocid)
		assert.Equal(t, "from-oci", res.DisplayName)
	}
	assert.Equal(t, 1, called, "expected single network invocation")
}

func TestResolveKmsKeyConfig_OCIGetKey_ServerError(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	keyID := testValidKeyOCID
	endpoint, err := getKmsEndpointFromKeyId(keyID)
	assert.NoError(t, err)

	called := 0

	p.kmsClientCache = make(map[string]oci.KmsClient)
	fk := &fakes.FakeKms{OnGet: func(ctx context.Context, r ocikms.GetKeyRequest) (ocikms.GetKeyResponse, error) {
		called++
		return ocikms.GetKeyResponse{}, fmt.Errorf("boom")
	}}
	p.kmsClientCache[endpoint] = fk

	res, err := p.ResolveKmsKeyConfig(ctx, &ociv1beta1.KmsKeyConfig{KmsKeyId: &keyID})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, 1, called)
}

func TestProvider_ResolveKmsKeyConfig_ContextCancellation(t *testing.T) {
	p := newTestProvider(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	keyID := testValidKeyOCID
	endpoint, _ := getKmsEndpointFromKeyId(keyID)
	p.kmsClientCache = make(map[string]oci.KmsClient)
	fk := &fakes.FakeKms{OnGet: func(ctx context.Context, r ocikms.GetKeyRequest) (ocikms.GetKeyResponse, error) {
		<-ctx.Done() // Wait for cancellation
		return ocikms.GetKeyResponse{}, ctx.Err()
	}}
	p.kmsClientCache[endpoint] = fk

	res, err := p.ResolveKmsKeyConfig(ctx, &ociv1beta1.KmsKeyConfig{KmsKeyId: &keyID})
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestProvider_GetKmsEndpointFromKeyId_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		ocid    string
		wantErr bool
	}{
		{"mixed-case vault", "ocid1.key.oc1.IAD.VaultA.ext1.ext2", false},
		{"unsupported region", "ocid1.key.oc1.xyz.vaultX.ext1.ext2", false}, // StringToRegion may handle unknown
		{"long extensions", "ocid1.key.oc1.iad.vaultA.ext1.ext2.ext3.ext4.ext5.ext6", false},
		{"short ocid", "ocid1.key.oc1.iad", true}, // Too few parts
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getKmsEndpointFromKeyId(tt.ocid)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvider_GetExtensions_Boundary(t *testing.T) {
	tests := []struct {
		name    string
		ocid    string
		wantErr bool
	}{
		{"5 parts", "ocid1.key.oc1.iad.vaultA", true},
		{"6 parts", "ocid1.key.oc1.iad.vaultA.ext1", false},
		{"7 parts", "ocid1.key.oc1.iad.vaultA.ext1.ext2", false},
		{"empty ocid", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getExtensions(tt.ocid)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvider_GetKmsClient_Concurrency(t *testing.T) {
	p := newTestProvider(t)
	endpoint, _ := getKmsEndpointFromKeyId("ocid1.key.oc1.iad.vaultX.ext1.ext2")

	// Seed cache to avoid client creation
	p.kmsClientCache = make(map[string]oci.KmsClient)
	fk := &fakes.FakeKms{}
	p.kmsClientCache[endpoint] = fk

	var wg sync.WaitGroup
	const numGoroutines = 50
	results := make([]*fakes.FakeKms, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			client, err := p.getKmsClient(endpoint)
			require.NoError(t, err)
			results[idx] = client.(*fakes.FakeKms)
		}(i)
	}
	wg.Wait()

	// All should be the same instance
	first := results[0]
	for _, r := range results {
		assert.Same(t, first, r)
	}
	assert.Equal(t, 0, first.GetCount.Get()) // No GetKey calls
}

func TestProvider_GetKmsClient_SDKError(t *testing.T) {
	p := newTestProvider(t)
	endpoint, _ := getKmsEndpointFromKeyId("ocid1.key.oc1.iad.vaultX.ext1.ext2")

	// Simulate SDK client construction failure by disabling keymanagement globally
	prev := common.OciSdkEnabledServicesMap
	t.Cleanup(func() { common.OciSdkEnabledServicesMap = prev })
	common.OciSdkEnabledServicesMap = map[string]bool{} // Disable all

	_, err := p.getKmsClient(endpoint)
	require.Error(t, err)
	// Just check that there's an error, don't assert message since it may vary
}
