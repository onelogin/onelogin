#!/usr/bin/env bash

VERSION=$(onelogin version)
LATEST_GIT_TAG=$(git describe --abbrev=0 --tags)
if [ $VERSION != $LATEST_GIT_TAG ]; then exit 1; fi

package=$1
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi
package_split=(${package//\// })
package_name=${package_split[${#package_split[@]}-1]}

platforms=("windows/amd64" "windows/386" "darwin/amd64" "linux/amd64" "linux/386")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='build/'$GOOS'-'$GOARCH'/'$package_name'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi

    cp './README.md' './build/'$GOOS'-'$GOARCH'/README.md'
    cp './LICENSE' './build/'$GOOS'-'$GOARCH'/LICENSE'

    if [ $GOOS = "windows" ]; then
      zip -r 'build/'$GOOS'-'$GOARCH'.zip' 'build/'$GOOS'-'$GOARCH
    else
      tar -czvf 'build/'$GOOS'-'$GOARCH'.tar.gz' 'build/'$GOOS'-'$GOARCH
    fi
    rm -rf 'build/'$GOOS'-'$GOARCH
done
