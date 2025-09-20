call ..\git_token.cmd
rem git tag -a v0.1.2 -m "First release"
rem git push origin v0.1.2
goreleaser release --clean