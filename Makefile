.PHONY: test web-build build build-linux verify

test:
	pwsh -File scripts/build.ps1 test
web-build:
	pwsh -File scripts/build.ps1 web-build
build:
	pwsh -File scripts/build.ps1 build
build-linux:
	pwsh -File scripts/build.ps1 build-linux
verify:
	pwsh -File scripts/build.ps1 verify
