#!/usr/bin/env sh

# Determine the platform using 'OS' and 'ARCH'
initPlatform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    if [ "$OS" = "darwin" ] && [ "$ARCH" = "arm64" ]; then
        PLATFORM="darwin_arm64"
    elif [ "$OS" = "darwin" ] && [ "$ARCH" = "x86_64" ]; then
        PLATFORM="darwin_amd64"
    elif [ "$OS" = "linux" ] && [ "$ARCH" = "x86_64" ]; then
        PLATFORM="linux_amd64"
    elif [ "$OS" = "linux" ] && { [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; }; then
        PLATFORM="linux_arm64"
    else
        printf "Platform not supported: os=%s arch=%s\n" "$OS" "$ARCH"
        exit 1
    fi
}

# Create a temporary directory and setup deletion on script exit using the 'EXIT' signal
initTmpDir() {
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf -- "${TMP_DIR}"' EXIT
    cd "$TMP_DIR"
}

# Ensure that dependency is installed and executable, exit and print help message if not
checkDependency() {
    if ! [ -x "$(command -v "$1")" ]; then
        printf "'%s' could not be found, this script depends on it, please install and try again.\n" "$1"
        exit 1
    fi
}

# Download the binary and check the integrity using the SHA256 checksum
downloadBinary() {
    CDN="cdn.rilldata.com"

    LATEST_URL="https://${CDN}/rill/latest.txt"
    if [ "${VERSION}" = "latest" ]; then
        VERSION=$(curl --silent --show-error ${LATEST_URL})
    fi
    BINARY_URL="https://${CDN}/rill/${VERSION}/rill_${PLATFORM}.zip"
    CHECKSUM_URL="https://${CDN}/rill/${VERSION}/checksums.txt"

    printf "Downloading binary: %s\n" "$BINARY_URL"
    curl --location --progress-bar "${BINARY_URL}" --output rill_${PLATFORM}.zip

    printf "\nDownloading checksum: %s\n" "$CHECKSUM_URL"
    curl --location --progress-bar "${CHECKSUM_URL}" --output checksums.txt

    printf "\nVerifying the SHA256 checksum of the downloaded binary:\n"
    shasum --algorithm 256 --ignore-missing --check checksums.txt

    printf "\nUnpacking rill_%s.zip\n" "$PLATFORM"
    unzip -q rill_${PLATFORM}.zip
}

# Print install options
printInstallOptions() {
    printf "\nWhere would you like to install rill?  (Default [1])\n\n"
    printf "[1]  /usr/local/bin/rill  [recommended, but requires sudo privileges]\n"
    printf "[2]  ~/.rill/rill         [directory will be created & path configured]\n"
    printf "[3]  ./rill               [download to the current directory]\n\n"
}

# Ask for preferred install option
promtInstallChoice() {
    printf "Pick install option: (1/2/3)\n"
    read -r ans </dev/tty;
    case $ans in
        1|"")
            INSTALL_DIR="/usr/local/bin"
            ;;
        2)
            INSTALL_DIR="$HOME/.rill"
            ;;
        3)
            INSTALL_DIR=$(pwd)
            ;;
        *)
            printf "\nInvalid option '%s'\n\n" "$ans"
            promtInstallChoice
            ;;
    esac
}

# Detect previous installation
detectPreviousInstallation() {
    if [ -x "$(command -v rill)" ] && [ -z "${INSTALL_DIR}" ]; then
        INSTALLED_RILL="$(command -v rill)"
        if [ "$INSTALLED_RILL" = "/usr/local/bin/rill" ]; then
            INSTALL_DIR="/usr/local/bin"
        elif [ "$INSTALLED_RILL" = "$HOME/.rill/rill" ]; then
            INSTALL_DIR="$HOME/.rill"
        fi
    fi
}

# Check conflicting installation and exit with a help message
checkConflictingInstallation() {
    if [ -x "$(command -v rill)" ]; then
        INSTALLED_RILL="$(command -v rill)"
        if [ -x "$(command -v brew)" ] && brew list rilldata/tap/rill >/dev/null 2>&1; then
            printf "There is a conflicting version of Rill installed using Brew.\n\n"
            printf "To upgrade using Brew, run 'brew upgrade rilldata/tap/rill'.\n\n"
            printf "To use this script to install Rill, run 'brew remove rilldata/tap/rill' to remove the conflicting version and try again.\n"
            exit 1
        elif [ "$INSTALLED_RILL" != "${INSTALL_DIR}/rill" ]; then
            printf "There is a conflicting version of Rill installed at '%s'\n\n" "$INSTALLED_RILL"
            printf "To use this script to install Rill, remove the conflicting version and try again.\n"
            exit 1
        fi
    fi
}

# Install the binary and ask for elevated permissions if needed
installBinary() {
    if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        printf "\nElevated permissions required to install the Rill binary to: %s/rill\n" "$INSTALL_DIR"
        sudo install -d "$INSTALL_DIR"
        sudo install rill "$INSTALL_DIR"
    else
        install -d "$INSTALL_DIR"
        install rill "$INSTALL_DIR"
    fi
    cd - > /dev/null
}

# Run the installed binary and print the version
testInstalledBinary() {
    RILL_VERSION=$("$INSTALL_DIR"/rill version)
    "$INSTALL_DIR"/rill verify-install 1>/dev/null || true
    printf "\nInstallation of %s completed!\n" "$RILL_VERSION"
}

# Print 'rill start' help intrcutions
printStartHelp() {
    boldon=$(tput smso)
    boldoff=$(tput rmso)

    if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        printf "\nTo start a new project in Rill, execute the command:\n\n %srill start my-rill-project%s\n\n" "$boldon" "$boldoff"
    elif [ "$INSTALL_DIR" = "$HOME/.rill" ]; then
        printf "\nTo start a new project in Rill, open a %snew terminal%s and execute the command:\n\n %srill start my-rill-project%s\n\n" "$boldon" "$boldoff" "$boldon" "$boldoff"
    elif [ "$INSTALL_DIR" = "$(pwd)" ]; then
        printf "\nTo start a new project in Rill, execute the command:\n\n %s./rill start my-rill-project%s\n\n" "$boldon" "$boldoff"
    fi
}

# Publish Syft install telemetry event, can be disabled by setting the 'RILL_INSTALL_DISABLE_TELEMETRY' environment variable
publishSyftEvent() {
    SYFT_URL=https://event.syftdata.com/log
    SYFT_ID=clp76quhs0006l908bux79l4v
    if [ -z "$RILL_INSTALL_DISABLE_TELEMETRY" ]; then
        curl --silent --show-error --header "Authorization: ${SYFT_ID}" --header "Content-Type: application/json" --data "{\"event_name\":\"$1\"}" $SYFT_URL > /dev/null || true >&2
    fi
}

# Add the Rill binary to the PATH via configuration of the shells we detect on the system
addPathConfigEntries() {
    PATH_CONFIG_LINE="export PATH=\$HOME/.rill:\$PATH # Added by Rill install"

    if [ "$INSTALL_DIR" = "$HOME/.rill" ]; then
        for f in "$HOME/.bashrc" "$HOME/.zshrc"; do
            if [ -f "$f" ]; then
                if ! grep -Fxq "$PATH_CONFIG_LINE" "$f"; then
                    printf "\nWould you like to add 'rill' to your PATH by adding an entry in '%s'? (Y/n)\n" "$f"
                    read -r ans </dev/tty;
                    case $ans in
                        n)
                            ;;
                        *)
                            printf "\n%s\n" "$PATH_CONFIG_LINE" >> "$f"
                            ;;
                    esac
                fi
            fi
        done
    fi
}

# Remove PATH configurations, we have to do handle this slightly different based on OS because of platform variations in 'sed' behaviour
removePathConfigEntries() {
    for f in "$HOME/.bashrc" "$HOME/.zshrc"; do
        if [ -f "$f" ]; then
            if [ "$OS" = "darwin" ]; then
                sed -i "" -e '/# Added by Rill install/d' "$f"
            elif [ "$OS" = "linux" ]; then
                sed -i -e '/# Added by Rill install/d' "$f"
            fi
        fi
    done
}

# Install Rill on the system
installRill() {
    publishSyftEvent install
    checkDependency curl
    checkDependency shasum
    checkDependency unzip
    initPlatform
    detectPreviousInstallation
    if [ -z "${INSTALL_DIR}" ]; then
        printInstallOptions
        promtInstallChoice
        checkConflictingInstallation
    fi
    initTmpDir
    downloadBinary
    installBinary
    testInstalledBinary
    addPathConfigEntries
    printStartHelp
    publishSyftEvent installed
}

# Uninstall Rill from the system, this function is aware of both the privileged and unprivileged install methods
uninstallRill() {
    checkDependency sed
    initPlatform

    if [ -f "/usr/local/bin/rill" ]
    then
        printf "\nElevated permissions required to uninstall the Rill binary from: '/usr/local/bin/rill'\n"
        sudo rm /usr/local/bin/rill
    fi

    rm -f "$HOME/.rill/rill"
    removePathConfigEntries

    printf "Uninstall of Rill completed\n"
}

set -e

# Parse input flag
case $1 in
    --uninstall)
        uninstallRill
        ;;
    --nightly)
        VERSION=nightly
        installRill
        ;;
    --version)
        VERSION=${2:-latest}
        installRill
        ;;
    --non-interactive)
        INSTALL_DIR=${2:-"/usr/local/bin"}
        VERSION=${3:-latest}
        installRill
        ;;
    *)
        VERSION=latest
        installRill
        ;;
esac
