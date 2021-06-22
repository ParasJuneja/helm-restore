# Go to the plugin dir
cd "$HELM_PLUGIN_DIR" || exit

# Fetch the version
version="$(grep "version" plugin.yaml | cut -d '"' -f 2)"

# set the url of the tar.gz
url="https://github.com/ParasJuneja/helm-restore/release/download/v${version}/helm-restore_${version}.tar.gz"

# set the filename
filename=$(echo "${url}" | sed -e "s/^.*\///g")

# download the archive using curl or wget
if [ -n "$(command -v curl)" ]
then
    curl -sSL -O "$url"
elif [ -n "$(command -v wget)" ]
then
    wget -q "$url"
else
    echo "Need curl or wget"
    exit 1
fi

# extract the plugin binary into the bin dir
rm -rf bin && mkdir bin && tar xzvf "$filename" -C bin > /dev/null && rm -f "$filename"

# Go back to original directory
pushd || exit