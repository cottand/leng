# Docker

Leng is also distributed as a Docker image.
[You can find published images here](https://github.com/Cottand/leng/pkgs/container/leng).
The image is small (v1.3.1 is under 13MB).

Supported architectures are linux AMD64, ARM64, ARMv6, ARMv7.

> If you think leng ought to support more OSs or architectures, please
[make an issue](https://github.com/Cottand/leng/issues/new).

## Running

With the default configuration:

```bash
docker run -d \
  -p 53:53/udp \
  -p 53:53/tcp \
  -p 8080:8080/tcp \
  ghcr.io/cottand/leng
```

With a specific `leng.toml`:

```bash
docker run -d \
  -p 53:53/udp \
  -p 53:53/tcp \
  -p 8080:8080/tcp \
  -v leng.toml:/leng.toml \
  ghcr.io/cottand/leng \
  -config /leng.toml

```

See [Configuratin](./Configuration.md) for the full config.