# Gil

```
gil price \
    --provider aws \
    --price-region "sa-east-1" \
    --label-selector "squad=psm-platform" \
    --exclude-containers 'istio-init,istio-proxy'
    # --all-containers
```

```
gil serve \
    --port 8080
```
