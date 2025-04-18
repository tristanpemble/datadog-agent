// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build windows

// Package utils contains common code shared across the USM codebase
package utils

import "net/http"

// GetTracedProgramsEndpoint is not supported on Windows
func GetTracedProgramsEndpoint(string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
	}
}

// GetBlockedPathIDEndpoint is not supported on Windows
func GetBlockedPathIDEndpoint(string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
	}
}

// GetClearBlockedEndpoint is not supported on Windows
func GetClearBlockedEndpoint(string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
	}
}

// GetAttachPIDEndpoint is not supported on Windows
func GetAttachPIDEndpoint(string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
	}
}

// GetDetachPIDEndpoint is not supported on Windows
func GetDetachPIDEndpoint(string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
	}
}
