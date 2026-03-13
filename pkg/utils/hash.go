/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func HashFor(any interface{}) (string, error) {
	// Marshal the struct to JSON
	jsonData, err := json.Marshal(any)
	if err != nil {
		return "", err
	}

	return Digest(jsonData), nil
}

func HashForMultiObjects(anys []interface{}) (string, error) {
	var bytes []byte

	for _, inf := range anys {
		if inf == nil {
			continue
		}

		jsonData, err := json.Marshal(inf)
		if err != nil {
			return "", err
		}

		bytes = append(bytes, jsonData...)
	}

	return Digest(bytes), nil
}

func Digest(b []byte) string {
	// Generate SHA256 hash
	hash := sha256.Sum256(b)

	// Convert the hash to a hex string
	return fmt.Sprintf("%x", hash)
}
