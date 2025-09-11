<!--
SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
SPDX-License-Identifier: Apache-2.0
-->

# go-api-declarations

[![CI](https://github.com/sapcc/go-api-declarations/actions/workflows/ci.yaml/badge.svg)](https://github.com/sapcc/go-api-declarations/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sapcc/go-api-declarations.svg)](https://pkg.go.dev/github.com/sapcc/go-api-declarations)

This Go module contains reusable declarations for types appearing in our APIs.
This repository is designed to have as little dependencies as possible and
contain as little application logic as possible. Also, by using versioned tags,
we avoid excessive auto-updates for this dependency in downstream repositories.
(This is in contrast to most of our Go repositories, which usually use
continuous delivery and do not tag releases.)
