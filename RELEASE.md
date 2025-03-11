# Releasing

This project uses [GoReleaser](https://goreleaser.com/) to create and publish releases.

To release a new version:

1. Bump the version in `spin-pluginify.toml`
2. Create a new tag:
   ```
   git tag -a v0.1.0 -m "Release v0.1.0"
   ```

3. Push the tag:
   ```
   git push origin v0.1.0
   ```
