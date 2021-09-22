# grpcurl Container

This container is useful for debugging gRPC in-cluster.

## Build/Push Image

```
bazel run //third_party/grpcurl:image_push
```

## Fire up container in namespace

```
KUBERNETES_NAMESPCE=funhouse
kubectl 
  -n $KUBERNETES_NAMESPACE \
  run \
  -i \
  -t \
  --rm \
  grpcurl \
  --image=ghcr.io/minorhacks/grpcurl:main \
  --restart=Never \
  --image-pull-policy="Always" \
  --overrides='{  "spec": { "imagePullSecrets": [{"name": "ghcr-cred"}] } }' \
  --command \
  -- \
  sh
```