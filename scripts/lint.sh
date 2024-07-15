#!/bin/bash

if [[ -n $(gofmt -l . | grep -v 'vendor/' | grep -v '.private-repos/') ]]; then
	echo "Some files fail gofmt checks:"
	gofmt -l .
	exit 1
fi;
echo "Code passes gofmt checks"
for m in $(find . -name go.mod | grep -v 'vendor/' | grep -v '.private-repos/'); do
	pushd $(dirname $m) > /dev/null;
	packages=$(go list ./...);
	if ! go vet $packages; then
		echo "Some packages in '$(dirname $m)' fail go vet";
		exit 1
	fi;
	popd > /dev/null;
done
echo "Code passes go vet checks"
