# Funhouse

A collection of distorted Git mirrors

## Building/Running FUSE mirror

1. Have a Linux machine with Bazel installed

1. Start the server, passing a URL to a public GitHub repository:

   ```
   bazel run //server:funhouse_server -- \
     --grpc_port=8080 \
     --base_path=/tmp/funhouse_data \
     --repo_url=https://github.com/minorhacks/advent_2020 \
     --alsologtostderr \
     --v=1
   ```

1. Start the client, passing the address to the server, as well as the directory
   to mount to:

   ```
   bazel run //client -- \
     --mount_point=/tmp/funhouse \
     --server_addr=localhost:8080 \
     --alsologtostderr
   ```

1. List files in a particular commit: `ls -la
   /tmp/funhouse/commits/0802d5e6cee084a8f867c5406e46a3fca556bf4e`

1. Run a build from a particular commit:

   NOTE: Writes in-tree will fail with `EROFS` (read-only filesystem) so build
   tools must be configured to produce build artifacts out-of-tree

   ```
   # The example is a Rust repo, so with the proper Rust tooling installed:
   CARGO_TARGET_DIR=/tmp/rust_build_out cargo build
   ```
