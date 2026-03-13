/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultImdsBaseUri = "http://169.254.169.254"
)

var (
	ErrRegionNotFound = errors.New("region info not found in IMDS")
)

// Generic method to retrieve any IMDS value based on endpoint.
var getInstanceMetadata = func(baseUri, endpoint string) (int, []byte, error) {
	// curl -s http://169.254.169.254/opc/v2/vnics/ -H "Authorization: Bearer Oracle"
	baseUri = strings.TrimRight(strings.TrimSpace(baseUri), "/")
	endpoint = strings.TrimLeft(strings.TrimSpace(endpoint), "/")
	if len(baseUri) == 0 {
		baseUri = defaultImdsBaseUri
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", baseUri, endpoint), nil)
	if err != nil {
		return 0, []byte{}, err
	}

	req.Header.Set("Authorization", "Bearer Oracle")
	res, err := client.Do(req)
	if err != nil {
		return 0, []byte{}, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	return res.StatusCode, body, err
}

func GetRegion() (string, error) {
	var ociRegion string
	statusCode, body, err := getInstanceMetadata("", "/opc/v2/instance")
	// IMDS endpoint can return trash locally, so only parse it if there is no error and 200 response code.
	if err != nil || statusCode != 200 {
		var ok bool
		if ociRegion, ok = os.LookupEnv("OCI_REGION"); !ok {
			return "", errors.New("no region info available")
		}
	} else {
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return "", err
		}
		region, ok := data["region"]
		if !ok {
			return "", ErrRegionNotFound
		}
		ociRegion = region.(string)
	}
	return ociRegion, nil
}
