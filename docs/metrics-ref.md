

## Metrics Reference

This page details the metrics that the custom metrics adapter exposes by default. Many of the exposed metrics are created in this project's dependencies. Generating this doc is currently a manual process.

### List of Stable Kubernetes Metrics

Stable metrics observe strict API contracts and no labels can be added or removed from stable metrics during their lifetime.


#### **apiserver_current_inflight_requests**
Maximal number of currently used inflight request limit of this apiserver per request kind in last second.

- **Stability Level:** STABLE
- **Type:** Gauge
- **Labels:** 

    - request_kind

#### **apiserver_request_duration_seconds**
Response latency distribution in seconds for each verb, dry run value, group, version, resource, subresource, scope and component.

- **Stability Level:** STABLE
- **Type:** Histogram
- **Labels:** 

    - component
    - dry_run
    - group
    - resource
    - scope
    - subresource
    - verb
    - version

#### **apiserver_request_total**
Counter of apiserver requests broken out for each verb, dry run value, group, version, resource, scope, component, and HTTP response code.

- **Stability Level:** STABLE
- **Type:** Counter
- **Labels:** 

    - code
    - component
    - dry_run
    - group
    - resource
    - scope
    - subresource
    - verb
    - version

#### **apiserver_response_sizes**
Response size distribution in bytes for each group, version, verb, resource, subresource, scope and component.

- **Stability Level:** STABLE
- **Type:** Histogram
- **Labels:** 

    - component
    - group
    - resource
    - scope
    - subresource
    - verb
    - version


### List of Beta Kubernetes Metrics

Beta metrics observe a looser API contract than its stable counterparts. No labels can be removed from beta metrics during their lifetime, however, labels can be added while the metric is in the beta stage. This offers the assurance that beta metrics will honor existing dashboards and alerts, while allowing for amendments in the future.


#### **disabled_metrics_total**
The count of disabled metrics.

- **Stability Level:** BETA
- **Type:** Counter


#### **hidden_metrics_total**
The count of hidden metrics.

- **Stability Level:** BETA
- **Type:** Counter


#### **registered_metrics_total**
The count of registered metrics broken by stability level and deprecation version.

- **Stability Level:** BETA
- **Type:** Counter
- **Labels:** 

    - deprecated_version
    - stability_level


### List of Alpha Kubernetes Metrics

Alpha metrics do not have any API guarantees. These metrics must be used at your own risk, subsequent versions of Kubernetes may remove these metrics altogether, or mutate the API in such a way that breaks existing dashboards and alerts.


#### **aggregator_discovery_aggregation_count_total**
Counter of number of times discovery was aggregated

- **Stability Level:** ALPHA
- **Type:** Counter


#### **apiserver_audit_event_total**
Counter of audit events generated and sent to the audit backend.

- **Stability Level:** ALPHA
- **Type:** Counter


#### **apiserver_audit_requests_rejected_total**
Counter of apiserver requests rejected due to an error in audit logging backend.

- **Stability Level:** ALPHA
- **Type:** Counter


#### **apiserver_client_certificate_expiration_seconds**
Distribution of the remaining lifetime on the certificate used to authenticate a request.

- **Stability Level:** ALPHA
- **Type:** Histogram


#### **apiserver_delegated_authz_request_duration_seconds**
Request latency in seconds. Broken down by status code.

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - code

#### **apiserver_delegated_authz_request_total**
Number of HTTP requests partitioned by status code.

- **Stability Level:** ALPHA
- **Type:** Counter
- **Labels:** 

    - code

#### **apiserver_envelope_encryption_dek_cache_fill_percent**
Percent of the cache slots currently occupied by cached DEKs.

- **Stability Level:** ALPHA
- **Type:** Gauge


#### **apiserver_flowcontrol_read_vs_write_current_requests**
Observations, at the end of every nanosecond, of the number of requests (as a fraction of the relevant limit) waiting or in regular stage of execution

- **Stability Level:** ALPHA
- **Type:** TimingRatioHistogram
- **Labels:** 

    - phase
    - request_kind

#### **apiserver_flowcontrol_seat_fair_frac**
Fair fraction of server's concurrency to allocate to each priority level that can use it

- **Stability Level:** ALPHA
- **Type:** Gauge


#### **apiserver_request_filter_duration_seconds**
Request filter latency distribution in seconds, for each filter type

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - filter

#### **apiserver_request_sli_duration_seconds**
Response latency distribution (not counting webhook duration and priority & fairness queue wait times) in seconds for each verb, group, version, resource, subresource, scope and component.

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - component
    - group
    - resource
    - scope
    - subresource
    - verb
    - version

#### **apiserver_request_slo_duration_seconds**
Response latency distribution (not counting webhook duration and priority & fairness queue wait times) in seconds for each verb, group, version, resource, subresource, scope and component.

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - component
    - group
    - resource
    - scope
    - subresource
    - verb
    - version
- **Deprecated Versions:** 1.27.0
#### **apiserver_storage_data_key_generation_duration_seconds**
Latencies in seconds of data encryption key(DEK) generation operations.

- **Stability Level:** ALPHA
- **Type:** Histogram


#### **apiserver_storage_data_key_generation_failures_total**
Total number of failed data encryption key(DEK) generation operations.

- **Stability Level:** ALPHA
- **Type:** Counter


#### **apiserver_storage_envelope_transformation_cache_misses_total**
Total number of cache misses while accessing key decryption key(KEK).

- **Stability Level:** ALPHA
- **Type:** Counter


#### **apiserver_tls_handshake_errors_total**
Number of requests dropped with 'TLS handshake error from' error

- **Stability Level:** ALPHA
- **Type:** Counter


#### **apiserver_webhooks_x509_insecure_sha1_total**
Counts the number of requests to servers with insecure SHA1 signatures in their serving certificate OR the number of connection failures due to the insecure SHA1 signatures (either/or, based on the runtime environment)

- **Stability Level:** ALPHA
- **Type:** Counter


#### **apiserver_webhooks_x509_missing_san_total**
Counts the number of requests to servers missing SAN extension in their serving certificate OR the number of connection failures due to the lack of x509 certificate SAN extension missing (either/or, based on the runtime environment)

- **Stability Level:** ALPHA
- **Type:** Counter


#### **authenticated_user_requests**
Counter of authenticated requests broken out by username.

- **Stability Level:** ALPHA
- **Type:** Counter
- **Labels:** 

    - username

#### **authentication_attempts**
Counter of authenticated attempts.

- **Stability Level:** ALPHA
- **Type:** Counter
- **Labels:** 

    - result

#### **authentication_duration_seconds**
Authentication duration in seconds broken out by result.

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - result

#### **authorization_attempts_total**
Counter of authorization attempts broken down by result. It can be either 'allowed', 'denied', 'no-opinion' or 'error'.

- **Stability Level:** ALPHA
- **Type:** Counter
- **Labels:** 

    - result

#### **authorization_duration_seconds**
Authorization duration in seconds broken out by result.

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - result

#### **field_validation_request_duration_seconds**
Response latency distribution in seconds for each field validation value

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - field_validation

#### **metrics_apiserver_metric_freshness_seconds**
Freshness of metrics exported

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - group

#### **workqueue_adds_total**
Total number of adds handled by workqueue

- **Stability Level:** ALPHA
- **Type:** Counter
- **Labels:** 

    - name

#### **workqueue_depth**
Current depth of workqueue

- **Stability Level:** ALPHA
- **Type:** Gauge
- **Labels:** 

    - name

#### **workqueue_longest_running_processor_seconds**
How many seconds has the longest running processor for workqueue been running.

- **Stability Level:** ALPHA
- **Type:** Gauge
- **Labels:** 

    - name

#### **workqueue_queue_duration_seconds**
How long in seconds an item stays in workqueue before being requested.

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - name

#### **workqueue_retries_total**
Total number of retries handled by workqueue

- **Stability Level:** ALPHA
- **Type:** Counter
- **Labels:** 

    - name

#### **workqueue_unfinished_work_seconds**
How many seconds of work has done that is in progress and hasn't been observed by work_duration. Large values indicate stuck threads. One can deduce the number of stuck threads by observing the rate at which this increases.

- **Stability Level:** ALPHA
- **Type:** Gauge
- **Labels:** 

    - name

#### **workqueue_work_duration_seconds**
How long in seconds processing an item from workqueue takes.

- **Stability Level:** ALPHA
- **Type:** Histogram
- **Labels:** 

    - name


