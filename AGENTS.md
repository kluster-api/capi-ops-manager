# AGENTS.md

This file provides guidance to coding agents (e.g. Claude Code, claude.ai/code) when working with code in this repository.

## Repository purpose

Go module `go.klusters.dev/capi-ops-manager` — a Kubernetes operator that exposes a `ClusterOpsRequest` CRD for cluster-lifecycle operations against Cluster API (CAPI) managed clusters. A `ClusterOpsRequest` describes an operation (e.g. Kubernetes version upgrade, component update, machine-pool changes), and the controller carries it out against the matching CAPI control plane / machine pools, with cloud-provider-specific code paths for AWS, Azure, and GCP.

There is also a `CAPIVersion` CRD that catalogs supported CAPI/Kubernetes versions per cloud — the controller reads it when validating ops requests.

The produced binary is `capi-ops-manager`. The README is mostly Kubebuilder scaffold; treat this file as the source of truth.

## Architecture

- `cmd/capi-ops-manager/main.go` — entry point.
- `pkg/cmds/` — Cobra commands:
  - `root.go` — top-level command.
  - `operator.go` — long-running controller process.
  - `webhook.go` — admission webhook process (validation/defaulting for ops requests).
- `apis/` — two API groups (`API_GROUPS := catalog:v1alpha1 ops:v1alpha1`):
  - `apis/ops/v1alpha1/` — `ClusterOpsRequest` types, helpers, and `clusteropsrequest_webhook.go` admission logic.
  - `apis/catalog/v1alpha1/` — `CAPIVersion` types.
  - Each group has `groupversion_info.go`, `*_types.go` (hand-written) and `zz_generated.*.go` + `openapi_generated.go` (generated).
- `pkg/controller/ops/` — the heart of the reconciler:
  - `clusteropsrequest_controller.go` — the main reconciler.
  - `manager.go` — wiring (scheme, client factories).
  - `update_version.go`, `update_components.go` — operations against the workload cluster.
  - `machinepool.go` — machine-pool operations.
  - `aws.go`, `azure.go`, `gcp.go` — per-cloud branches called by the above (CAPA/CAPZ/CAPG).
  - `util.go` — shared helpers.
- `config/` — Kubebuilder Kustomize bases (CRDs, RBAC, manager, webhook).
- `crds/` — versioned CRD YAML manifests (`catalog.klusters.dev_capiversions.yaml`, `ops.klusters.dev_clusteropsrequests.yaml`) plus `lib.go` exposing them via `go:embed`.
- `PROJECT` — Kubebuilder project metadata. Do not hand-edit unless also using `kubebuilder edit`. Multigroup mode is enabled.
- `test/` — test harness (Ginkgo/Gomega).
- `Dockerfile.in` (PROD, distroless), `Dockerfile.dbg` (debian), `Dockerfile.ubi` (Red Hat certified) — three image variants per release.
- `hack/`, `Makefile` — AppsCode build harness (runs everything inside `ghcr.io/appscode/golang-dev`).
- `vendor/` — checked-in deps.

The repo predates the move to the AppsCode build harness — `main.go` and `cmd/` use the **Apache-2.0** header rather than the AppsCode Community License used in newer kluster.dev repos. Match the existing header style when adding files to a package.

## Common commands

All Make targets run inside `ghcr.io/appscode/golang-dev` — Docker must be running.

- `make ci` — CI pipeline: `check-license lint build` (note: `unit-tests`, `cover`, and `verify` are commented out in CI; run them locally before opening a PR).
- `make gen` — regenerate everything: `manifests openapi`. Run after any change to `apis/**/*_types.go`.
- `make manifests` — regenerate CRDs only.
- `make openapi` — regenerate OpenAPI definitions only.
- `make clientset` — regenerate client code (target exists but no `client/` dir is currently checked in).
- `make fmt` — gofmt + goimports.
- `make lint` — golangci-lint.
- `make unit-tests` — Go unit tests.
- `make e2e-tests` / `make test` — runs both unit and e2e.
- `make verify` — `verify-gen verify-modules`; `go mod tidy && go mod vendor` must leave the tree clean.
- `make container` — build PROD, DBG, and UBI images.
- `make push` — push all three; `make docker-manifest` writes multi-arch manifests; `make release` is the full publish flow.
- `make push-to-kind` / `make deploy-to-kind` — load into Kind and Helm-install.
- `make install` / `make uninstall` / `make purge` — Helm install lifecycle.
- `make add-license` / `make check-license` — manage license headers.

Run a single Go test (requires a local Go toolchain):

```
go test ./pkg/controller/ops/... -run TestName -v
```

## Conventions

- Module path is `go.klusters.dev/capi-ops-manager` (vanity URL); imports must use that.
- License: **Apache-2.0** (`LICENSE`) — note this differs from sibling kluster.dev operators which use the AppsCode Community License. Use the existing file's header style when adding files: Apache-2.0 in current packages.
- Sign off commits (`git commit -s`); contributions follow the DCO (project ships a `DCO` requirement via the Kubebuilder template).
- Vendor directory is checked in — `go mod tidy && go mod vendor` must leave the tree clean (enforced by `verify-modules`).
- Per-cloud code lives strictly in `pkg/controller/ops/{aws,azure,gcp}.go`. Don't sprinkle cloud-specific branches across the main reconciler — call into those files instead.
- Do not hand-edit `zz_generated.*.go`, `openapi_generated.go`, or any file under `crds/` — change `apis/**/*_types.go` and re-run `make gen`.
- Three Dockerfiles, one binary — keep `Dockerfile.in`, `Dockerfile.dbg`, and `Dockerfile.ubi` in sync.
- This is a **Kubebuilder multigroup project** (`PROJECT`): use `kubebuilder edit` rather than hand-modifying `PROJECT`; add new APIs with `kubebuilder create api`.
