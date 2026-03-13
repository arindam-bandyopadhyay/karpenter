/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintMapAndQuote[K comparable, V any](m map[K]V) string {
	return PrintMapAndQuoteWithControl(m, func(k K, v V) string {
		return fmt.Sprintf("%v=%v", k, v)
	})
}

func PrintSliceAndQuoteWithControl[V any](s []V, control func(v V, i int) string) string {
	return fmt.Sprintf("%s%s%s", "\"", strings.Join(lo.Map(s, control), ","), "\"")
}

func PrintMapAndQuoteWithControl[K comparable, V any](m map[K]V, control func(k K, v V) string) string {
	return fmt.Sprintf("%s%s%s", "\"", strings.Join(lo.MapToSlice(m, control), ","), "\"")
}

func PrintMapAndQuoteWithSingleQuote[K comparable, V any](m map[K]V, control func(k K, v V) string) string {
	return fmt.Sprintf("%s%s%s", "'", strings.Join(lo.MapToSlice(m, control), ","), "'")
}

func PrettyPrintAsJson(format string, data interface{}) string {
	b := lo.Must(json.MarshalIndent(data, "", "  "))
	return fmt.Sprintf(format, b)
}

func PrintMapAndQuoteDuration(m map[string]v1.Duration) string {
	return PrintMapAndQuoteWithControl(m, func(k string, v v1.Duration) string {
		return fmt.Sprintf("%v=%v", k, v.Duration)
	})
}

func PrintMapAndQuoteForEviction[K comparable, V any](m map[K]V) string {
	return PrintMapAndQuoteWithSingleQuote(m, func(k K, v V) string {
		return fmt.Sprintf("%v<%v", k, v)
	})
}

func PrintMapAndQuoteDurationForEviction(m map[string]v1.Duration) string {
	return PrintMapAndQuoteWithSingleQuote(m, func(k string, v v1.Duration) string {
		return fmt.Sprintf("%v=%v", k, v.Duration)
	})
}
