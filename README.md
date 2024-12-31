# promqlfmt

Code formatter for [PromQL].

Usage:

```
$ promqlfmt **/*.promql
promqlfmt: Formatted dashboards/my-service/cache-hits.promql
promqlfmt: Formatted dashboards/my-service/requests-total.promql
promqlfmt: Formatted dashboards/my-service/error-percentage.promql
```

Dry-run with word diff (default):

```console
$ promqlfmt -dry-run dashboards/my-service/cache-hits.promql
promqlfmt: Diffing dashboards/my-service/cache-hits.promql
diff --git a/dashboards/my-service/cache-hits.promql b/dev/fd/3
index 981f116fc64d..000000000000 100644
--- a/dashboards/my-service/cache-hits.promql
+++ b/dev/fd/3
@@ -1,3 +1,3 @@
sum by [-(shard_name,cache_hit)-]{+(shard_name, cache_hit)+} (
  [-rate(some_random_server_cache_operations_total{cache_name="address_lookup", result="hit"}[5m])-]{+rate(some_random_server_cache_operations_total{cache_name="address_lookup",result="hit"}[5m])+}
)
```

Dry-run with unified diff:

```console
$ promqlfmt -dry-run -differ=diff dashboard/my-service/cache-hits.promql 
promqlfmt: Diffing dashboard/my-service/cache-hits.promql
--- dashboard/my-service/cache-hits.promql    2024-12-31 18:30:53.730188541 +0000
+++ /dev/fd/3   2024-12-31 18:35:57.329929607 +0000
@@ -1,3 +1,3 @@
-sum by (shard_name,cache_hit) (
-    rate(some_random_server_cache_operations_total{cache_name="address_lookup", result="hit"}[5m])
+sum by (shard_name, cache_hit) (
+  rate(some_random_server_cache_operations_total{cache_name="address_lookup",result="hit"}[5m])
 )
```

Installation:

```
$ go install github.com/mikedanese/promqlfmt@latest
```

## Disclaimer

This is a trivial prototype that uses the parser from prometheus to parse the
query expression and spit out a prettified version. 

## TODO

* Probably only works on Unix. Someone could add windows support.
* This doesn't work well for parameterized queries (e.g. if you are
  interpolating variables). I may add support for such things, but will need to
  reimplement some of the parser.
* Since we have the AST, this would be a good place to add lints if we can think
  of any.


[PromQL]: https://prometheus.io/docs/prometheus/latest/querying/basics/