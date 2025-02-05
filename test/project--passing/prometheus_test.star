observability = import_module("github.com/ethpandaops/optimism-package/src/observability/observability.star")
prometheus_launcher = import_module("github.com/ethpandaops/optimism-package/src/observability/prometheus/prometheus_launcher.star")

def test_prometheus(plan):
    observability_params = struct(
        enabled=True,
        prometheus_params=struct(
            storage_tsdb_retention_time="1d",
            storage_tsdb_retention_size="512MB",
            min_cpu=10,
            max_cpu=1000,
            min_mem=128,
            max_mem=2048,
            image="prom/prometheus:latest",
        )
    )
    observability_helpers = observability.make_helper(observability_params)
    global_node_selectors = {}

    prometheus_url = prometheus_launcher.launch_prometheus(plan, observability_helpers, global_node_selectors)
    prometheus = plan.get_service(name = "prometheus")
    prometheus_ip_address = prometheus.ip_address
    prometheus_http_port = prometheus.ports["http"].number
    expected_prometheus_url = "http://{0}:{1}".format(
        prometheus_ip_address, prometheus_http_port
    )

    assert.eq(prometheus_url, expected_prometheus_url)