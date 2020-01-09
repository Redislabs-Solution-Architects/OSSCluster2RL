#!/usr/bin/env bash

rm -rf bin/osscluster2rl_*

package="osscluster2rl.go"
package_name="osscluster2rl"

platforms=("windows/amd64" "linux/amd64" "darwin/amd64" "linux/386" "windows/386")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'_'$GOOS'_'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi  

    env GOOS=$GOOS GOARCH=$GOARCH go build -o bin/${output_name} $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done

mkdir -p /tmp/osscluster2rl
cp -R bin/* /tmp/osscluster2rl
cp README.md /tmp/osscluster2rl
cp example_config.toml /tmp/osscluster2rl
cd /tmp
tar -cf osscluster2rl.tar osscluster2rl
gzip osscluster2rl.tar
