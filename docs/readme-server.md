---
authors: Mark Epstein (epstein137e@gmail.com)
state: draft
---

# HTTP API for Kubernetes Resources

## What

This project is intended to specify, code, document, and package a
server that will provide an API that allows access to Kubernetes resource
values via HTTP.

## Why

### Technical
We need a simple way of accessing Kubernetes resource values via HTTP.

### Process
Standard RFD process.

## Required Approvers
Two of the below:

- Evan Freed (Github: evanfreed)
- Logan Davis (Github: logand22)
- Stephen Levine (Github: sclevine)
- Jim Bishopp (Github: jimbishopp)
- Russ Jones (Github: russjones)

## Details

### Requirements
- Written concisely and clearly
- Well-documented
- Easy to use

#### Server
- Written in go
- Able to be run both inside and outside a Kubernetes cluster
- Built via Make
- Packaged into a Docker container
- Ideally deployed via Helm

#### API
- RESTful
- Available via TLS
- Uses a standard authentication mechanism

### Background
This is code I wrote to make Kubernetes configuration values more
accessible.

### API

#### General
- Requests are made in RESTful format.
- Responses contain a JSON reply body, which may be empty.

#### Requirements
- Cache necessary info by watching for changes to Deployments
- Read-only requests should not each trigger a request to the cluster.
Note: It is acceptable to use either client-go or controller-runtime
to implement this.

#### Per-Endpoint Detail
| ID | Description | Type | Arguments | Request Schema | Example | Example Response |
| -- | :----------- | ----------- | :----------- | :----------- | :----------- | :----------- |
| 1  | List namespaces in the cluster | GET | \<none\> | /namespaces | /namespaces | ```{ "namespaces": [ "kube-system", "personal" ] }``` |
| 2  | List deployments in a namespace | GET | namespace | /namespaces &nbsp;&nbsp;/:namespace &nbsp;&nbsp;/deployments | /namespaces &nbsp;&nbsp;/*personal* &nbsp;&nbsp;/deployments | ```{ "namespace": "personal", "deployments": [ "nginx", "kafka", "resource-access" ] }``` |
| 2A | List deployments in all namespaces | GET | \<none\> | /namespaces &nbsp;&nbsp;/ANY &nbsp;&nbsp;/deployments | /namespaces &nbsp;&nbsp;/*ANY* &nbsp;&nbsp;/deployments | ```[ { "namespace": "personal", "deployments": [ "nginx", "kafka", "resource-access" ] }, { "namespace": "kube-system", "deployments": [ "fred", "jane", sally" ] } ]``` |
| 3  | Get deployment replica count | GET | namespace deployment | /namespaces &nbsp;&nbsp;/:namespace &nbsp;&nbsp;/deployments &nbsp;&nbsp;/:deployment &nbsp;&nbsp;/replica\_count | /namespaces &nbsp;&nbsp;/*personal* &nbsp;&nbsp;/deployments &nbsp;&nbsp;/*nginx* &nbsp;&nbsp;/replica\_count | ```{ "namespace": "personal", "deployment": "nginx", "replica_count": 12 }``` |
| 3A | Get all deployment replica counts for a namespace | GET | namespace | /namespaces &nbsp;&nbsp;/:namespace &nbsp;&nbsp;/deployments &nbsp;&nbsp;/ANY &nbsp;&nbsp;/replica\_count | /namespaces &nbsp;&nbsp;/*personal* &nbsp;&nbsp;/deployments &nbsp;&nbsp;/ANY &nbsp;&nbsp;/replica\_count | ```{ "namespace": "personal", "deployments": [ { "deployment": "nginx", "replica_count": 12 }, { "deployment": "server", "replica_count": 3 } ] }``` |
| 4  | Set deployment replica count | PUT | namespace deployment replica\_count | /namespaces &nbsp;&nbsp;/:namespace &nbsp;&nbsp;/deployments &nbsp;&nbsp;/:deployment &nbsp;&nbsp;/replica\_count &nbsp;&nbsp;/:replica\_count | /namespaces &nbsp;&nbsp;/*personal* &nbsp;&nbsp;/deployments &nbsp;&nbsp;/*nginx* &nbsp;&nbsp;/replica\_count &nbsp;&nbsp;/*38* | |
| 5  | Get *liveness* state | GET | \<none\> | /livez | /livez | |
| 6  | Get *readiness* state | GET | \<none\> | /readyz | /readyz | |


#### HTTP Status Codes

##### Implemented
These status codes are implemented in this version of the service.
| Code | Title | Endpoint(s) | Detail/Notes |
| ---- | :------ | ----------- | :----------- |
| 200  | OK | [all] | | For endpoint 4, this status code indicates that the number of replicas has been successfully set. |
| 400  | Bad Request | [all] | Syntax error in request, or similar. |
| 401  | Unauthorized | 1-4 | Identified user does not have permission to perform this action. |
| 403  | Forbidden | 1-4 | User has not been identified. |
| 404  | Not Found | [unidentified] | Unknown endpoint. |
| 410  | Gone | 1-4 | Resource requested no longer exists. |
| 503  | Service Unavailable | [all] | For endpoints 5 and 6, this code indicates "not live / not ready". |

##### Unimplemented
These status codes would be implemented in a *real* version of the service.
| Code | Title | Endpoint(s) | Detail/Notes |
| ---- | :------ | ----------- | :----------- |
| 408  | Request Timeout | 1-4 | Server took too long to process user request. |
| 414  | URI Too Long | [all] | |
| 429  | Too Many Requests | [all] | Implemented to prevent denial of service attacks, whether accidental or intentional. |
| 431  | Request Header Fields Too Large | [all] | |
| 500  | Internal Server Error | [all] | |


### Developer Workflow

#### How to Make Changes
- Clone the repo
- Make, document and commit your changes
- Run unit and integration tests
- Create a PR that encapsulates the changes
- Iterate on the PR with approvers
- Merge

Ideally once the merge is complete, CI will create a reproducible
Docker container and place it such that it can be automatically
retrieved by Kubernetes.

#### How to Release and Deploy

##### General
This service stores no significant state, and what state there is can
be regenerated quickly, easily, and at a fairly low cost. Plus, the
service is entirely stateless from the user point of view.  So to
deploy a new version without any service outage:<br>
- Generate a new artifact and place it such that it can be automatically fetched
- Perform a rolling restart of the service

Were a short service outage acceptable, version upgrade could be simplified to:<br>
- Generate a new artifact and place it such that it can be automatically fetched
- Do a `helm uninstall` of the service, followed by a `helm install`


##### First-Time Installation in a Cluster
- Clone the repo: `git clone https://github.com/fastiterator/probable-potato`.
- cd to `probable-potato/build`, then do `make help` to see the make targets.
- Set env var `DH_USER` to your Docker Hub username.
- Do `make push_docker` - this will build the server and push it to Docker Hub.
- Set env var `AWS_REGION` to the AWS region in which you want your
  EKS cluster built. This value defaults to `us-west-2`.
- You will need to have or set up a variety of AWS permissions,
  including those for EKS. Please see `docs/iam.md` in this repo for
  details.
- Do `make create_cluster` to create an EKS cluster for this service to
  run in. This _should_ also set up kubectl to point at your new cluster.
- Do `helm install server` to install the server in the cluster.

##### Version Upgrade
- Check out the cluster spec / config and/or helm chart for the service
- Update the config  to reflect the desired service version.
- PR & merge the config change.
- If not, do a rolling restart of the service in the cluster using
  agreed / approved methods.

### Build Process
- To build locally, use `Make`.
- CI will use `Make` as well, and will then deliver the built artifact
  to an artifact repo, from which Kubernetes can automatically fetch
  it.

### Release Process
Please see `How to Release and Deploy`, above.

### Caching

#### Prototype / Light Load
This initial version of this server will use per-pod in-memory
caching, because it is simple and has very few moving parts.

#### Heavy Load
Were this server to need to operate at very large scale, and/or were
it critical for it to make as little impact to Kubernetes as possible,
I would recommend using replicated `redis` as a caching layer. The
benefits of using this sort of persistent and centralized caching
layer are as follows:
- Instead of each pod needing to make a request to Kubernetes for each
  data object, it would be each deployment. In the case where there
  are many pods, this difference can matter.
- On startup, a new pod (other than the first pod to start), will have
  access to a fully populated cache, rather than starting with a
  cold/slow cache. This allows it to serve useful traffic sooner, and
  with less likelihood of being buried by the initial onslaught of
  requests.
- Having a persistent, centralized cache allows us to break the
  listen-and-update functionality completely apart from the service of
  client requests, which should lead to simpler and more maintainable
  code.

### Scaling
Were this a *real* service, it would be configured to auto-scale
itself based on usage load. Since the service is to be run in a
Kubernetes cluster, the most straightforward way to do this would be
to set up Kubernetes' `Horizontal Pod Autoscaling` (HPA) system for
this service. The configuration would instruct HPA to periodically
poll the service's pods for CPU use, and to add a pod when current
pods' average CPU was over e.g. 40%.

### mTLS

#### Description
Typical best practice for RESTful APIs is to accept only TLS-encrypted
connections. These allow clients to authenticate the server, and
prevent many types of security issues. This service will run according
to Zero Trust via use of Mutual TLS (mTLS).  The server will present a
certificate to the client, enabling the client to authenticate the
server, and the client will do the same for the server. This mutual
(two-way) authentication arrangement prevents even more security
issues than the one-way authentication in common use on the public
Internet, but at the cost of higher management overhead.

#### General
Since this service uses mTLS, neither client nor server can use
certificates signed by a public certificate authority (CA). Instead,
the company needs to run its own CA. Client and server will have
certificates generated from the same root CA. Both client and server
will have certificates to use at runtime that the other will
recognize, otherwise they would be unable to communicate with each
other.

#### Details

##### TLS Version
TLS 1.2 is widely used, but has some performance issues and also some
security issues, both of which classes of issues were solved by the
introduction of TLS 1.3. Since we control both ends of this
client/server connection, we have full freedom of choice with respect
to TLS version, so TLS 1.3 is what we will use.

##### Cipher Suite
The list of cipher suites permissible for use with TLS 1.3, and also
supported by OpenSSL, the software we will use to interface with
certificates, is very short.

*`% openssl ciphers -v -V -s -tls1_3 'ALL:@STRENGTH'`*<br>
```TLS_AES_256_GCM_SHA384         TLSv1.3 Kx=any      Au=any   Enc=AESGCM(256)            Mac=AEAD```<br>
```TLS_CHACHA20_POLY1305_SHA256   TLSv1.3 Kx=any      Au=any   Enc=CHACHA20/POLY1305(256) Mac=AEAD```<br>
```TLS_AES_128_GCM_SHA256         TLSv1.3 Kx=any      Au=any   Enc=AESGCM(128)            Mac=AEAD```<br>

The ``CHACHA`` cipher is faster in pure software, i.e. on hardware
that does not have AES-specific instructions. All modern Intel CPUs
used in AWS EC2 do support the AES instructions, so this is likely not
an issue.

Both `TLS_AES_128_GCM_SHA256` and `TLS_AES_256_GCM_SHA384` are
current, recommended cipher suites. The difference between the two is
just one of numbers, performance, and possibly regulatory
compliance. The former is weaker but also faster and less expensive;
the latter is stronger but slower and more expensive. Given:
- Both are currently recommended, and likely provide more than
  sufficient protection.
- There are no specific regulations that restrict us to use of the
  more expensive of these two.

So I plan to use `TLS_AES_128_GCM_SHA256` as the mTLS cipher suite for
this service.

### Delivery
Were this a *real* service, its artifacts woukld be delivered by CI or
human means to an artifact repo. From there, the service could/would
be fetched and installed onto a Kubernetes cluster with Helm.

### Security
There are many facets of security. This service is an API, so is
vulnerable only to a subset of vulnerabilities. The most likely
vulnerabilities and their mitigations are detailed below.

#### Attack Vectors & Mitigations
To reduce the possibility of successful attack, we will do the
following...

##### Attack Surface Area
Reduction of attack surface area is one of the single most valuable
and low-cost avenues for security breach reduction.
- Since this service is for company use only, and the company
  presumably uses private address space as much as reasonably
  practicable, this service need not--and should not--have an
  Internet-facing IP address. Were this a *real* service, it would be
  exposed only inside the company.
- Were this a *real* service, we would prevent use of this service by
  unauthorized entities by maintaining an authorization layer within
  the service.
- The service only opens ports that must be open to meet 
  service needs.

##### Externally-Sourced Code
Externally-sourced code and libraries have been the source of many
security failures.
- This service depends only on Go standard library components and
  standard Kubernetes components. This software is composed of
  frequently-scrutinized, regularly updated code, and any discovered
  security vulnerabilities are corrected quickly.
- Were this a *real* service, we would track security vulnerability
  status of components used by this service, and will updated as
  needed.

##### DDoS and Outage-Type Attacks
This service, like any API, could suffer from DDoS attacks.
- The primary mitigation against these attacks in this particular
  service is its deliberately low attack surface area.
- Counting use over short periods of time--and then limiting service
  use by by particular connections and/or accounts if their the usage
  level has been too high recently--can also help reduce the impact of
  these type of attacks. Were this a *real* service, we would most
  likely implement a token bucket type rate limiting system.

##### User-Sourced Information
Maliciously-crafted user-sourced information, that is not vetted or
cleaned by the service, can represent a danger of code injection,
among other issues. This service takes in almost no user-sourced
information, and what it does take in, it checks carefully before use.

##### Buffer Overflows
This service uses libraries that closely control their own buffers,
and the service is careful with its own buffers. Together, these
should help prevent buffer-overflow related exploits.

### UX
This service's API allows only a single usage path:
- User authentication & authorization
- Get list of namespaces in the cluster
- Get list of deployments for a namespace
- Get or set number of replicas for a deployment

There is no UI here, and no CLI here.

### Observability
Were this a *real* service, observability would be accomplished
through a variety of techniques and data sources. Please see below for
details.

#### Dashboards
Were this a *real* service, we would create a Grafana dashboard for
this service that details service-related information, including but
not limited to accessibility and performance.

#### Service-Emitted Metrics
Were this a *real* service, it would emit metrics to an observability
platform such as Prometheus. The minimal metrics to be emitted would
be as follows

- service startup
- service shutdown
- instance startup: instance
- instance shutdown: instance
- incoming-request: instance, req-id, req-type, req-args
- completed request: instance, req-id, disposition
- info / warning / error: instance, type, content

#### Kubernetes service health metrics
Were this a *real* service, we would ensure that Kubernetes' service
health monitoring emit metrics to Prometheus sufficient to determine
the service's health.

#### Synthetic Queries
Were this a *real* service, we would set up periodic synthetic
queries, and use them to test the service end-to-end in the real world
on an ongoing basis. The synthetic queries would each emit a
success/fail metric.

#### Alerts
Were this a *real* service, we would set up alerts for:

- Inaccessibility (from synthetic queries)
- Poor health (from Kubernetes health metrics)
- Excessive error rate (from service-emitted metrics)
- p95 request completion time  (from service-emitted metrics)
- High rate of instance turnover (from a combination of Kubernetes
  service metrics and service-emitted metrics)

Together, these alerts would allow us to know and act on service poor
performance, including outright failure, without needing to look at
graphs or the like.

### Product Usage
This is specified to be an internal service that would be used by
other services. Were this a *real* service, we would determine
adoption via observation of user counts and usage levels, both of
which would be available as metrics and reflected on the Grafana
service dashboard.

### Test Plan
There are a set of unit and integration tests that are included with
this service package, at least some of which will be run by CI on code
check-in and/or branch merge, and all of which should be passed prior
to the code being released to production.

As new functionality is added to the service, new tests must be
written and integrated into the package.

As bugs are found, tests must be written which would expose any
service regression to the previously-seen bug-containing state.

### Notes

#### Process
- Per Github current standard practice, the repo in which this RFD
  resides uses `main` rather than `master` as its primary branch.

#### Documents
##### mTLS Resources
Step by step guide to mTLS in Go —
[https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go](https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go)
TLS_AES_128_GCM_SHA256 Cipher Suite — [https://ciphersuite.info/cs/TLS_AES_128_GCM_SHA256/](https://ciphersuite.info/cs/TLS_AES_128_GCM_SHA256/)
TLS_AES_256_GCM_SHA384 Cipher Suite — [https://ciphersuite.info/cs/TLS_AES_256_GCM_SHA384/](https://ciphersuite.info/cs/TLS_AES_256_GCM_SHA384/)
Short Explanation of TLS 1.3 Cipher Suites — [https://crypto.stackexchange.com/questions/63796/why-does-tls-1-3-support-two-ccm-variants](https://crypto.stackexchange.com/questions/63796/why-does-tls-1-3-support-two-ccm-variants)<br>
Grover's Algorithm — (https://en.wikipedia.org/wiki/Grover%27s_algorithm)[https://en.wikipedia.org/wiki/Grover%27s_algorithm]
