#!/bin/bash -e

attempt=1

usage="
Script starts or continues prebid upgrade to version set in 'to_minor' variable. Workspace is at {Workspace}/prebid-server and {Workspace}/pbs-patch

    ./upgrade-pbs.sh [--restart] [--version=VERSION] [--workspace=PATH]

    --restart   Restart the upgrade (deletes {Workspace}/prebid-server and {Workspace}/pbs-patch)
    --version=VERSION  Specify a particular version to upgrade to (optional)
    --workspace=PATH   Specify a workspace path (default: /GOPATH/src).
    -h          Help
TODO:
    - paramertrize the script
    - create ci branch PR
    - create header-bidding PR"
RESTART=0
VERSION=""
WORKSPACE="${GOPATH:-$(go env GOPATH)}/src"

# Process arguments
process_arguments() {
    for i in "$@"; do
        case $i in
            --restart)
                RESTART=1
                ;;
            -h)
                echo "$usage"
                exit 0
                ;;
            --version=*)
                VERSION="${i#*=}"
                ;;
            --workspace=*)
                WORKSPACE="${i#*=}"
                ;;    
        esac
    done
}

# Restart upgrade
restart_upgrade() {
    if [ "$RESTART" -eq "1" ]; then
        log "Restarting the upgrade: rm -rf $WORKSPACE/prebid-server $WORKSPACE/pbs-patch/"
        rm -rf $WORKSPACE/prebid-server $WORKSPACE/pbs-patch/
        mkdir -p $WORKSPACE/pbs-patch/
    fi
}

# Initialize upgrade
initialize_upgrade() {
    checkpoint_run clone_repo
    cd $WORKSPACE/prebid-server
    log "At $(pwd)"

    # Get the latest tag if VERSION is not specified
    if [ -z "$VERSION" ]; then
        VERSION=$(git describe --tags $(git rev-list --tags --max-count=1))
    fi

    log "Final Upgrade Version: $VERSION"

    git diff tags/$VERSION..origin/master > $WORKSPACE/pbs-patch/current_ow_patch-$VERSION-origin_master-$attempt.diff
}

# Start validation
start_validation() {
    go_mod
    checkpoint_run "./validate.sh --race 5 --nofmt"
    go_discard
}

# Setup branches
setup_branches() {
    tag_base_branch_name=prebid_$VERSION-$attempt-tag
    upgrade_branch_name=prebid_$VERSION-$attempt
    log "Reference tag branch: $tag_base_branch_name"
    log "Upgrade branch: $upgrade_branch_name"

    checkpoint_run checkout_branch
}

# Merge branches
merge_branches() {
    log "Merging master into $tag_base_branch_name"
    checkpoint_run git merge master --no-edit

    log "Validating the master merge into current tag. Fix and commit changes if required."
    go_mod
    checkpoint_run "./validate.sh --race 5 --nofmt"
    go_discard

    checkpoint_run git checkout master
    checkpoint_run git merge $upgrade_branch_name --no-edit
}

# Generate patch file
generate_patch() {
    log "Generating patch file at $WORKSPACE/pbs-patch/ for $VERSION"
    git diff tags/$VERSION..master > $WORKSPACE/pbs-patch/new_ow_patch_$VERSION-master-1.diff
}

# Log message
log() {
    printf "\n$(date): $1\n"
}

# Clear log on exit
clear_log() {
    current_fork_at_version=$(git describe --tags --abbrev=0)
    if [ "$current_fork_at_version" == "$VERSION" ]; then
        log "Upgraded to $current_fork_at_version"
        rm -f "$CHECKLOG"

        log "Last validation before creating PR"
        go_mod
        checkpoint_run "./validate.sh --race 5 --nofmt"
        go_discard

        set +e
        log "Commit final go.mod and go.sum"
        git commit go.mod go.sum --amend --no-edit
        set -e

     else
        log "Exiting with failure!!!"
        exit 1
    fi
}

# Clone repository
clone_repo() {
    if [ -d "$WORKSPACE/prebid-server" ]; then
        log "Code already cloned. Attempting to continue the upgrade!!!"
    else
        log "Cloning repo at $WORKSPACE"
        cd $WORKSPACE
        git clone https://github.com/PubMatic-OpenWrap/prebid-server.git
        cd prebid-server

        git remote add prebid-upstream https://github.com/prebid/prebid-server.git
        git fetch --all --tags --prune
    fi
}

# Checkout branch
checkout_branch() {
    set +e 
    git checkout tags/$VERSION -b $tag_base_branch_name
    git checkout -b $upgrade_branch_name
    git checkout $upgrade_branch_name
    set -e
}

# Execute command
cmd_exe() {
    cmd=$*
    if ! $cmd; then
        log "Failure!!! creating checkpoint $cmd"
        echo "$cmd" > $CHECKLOG
        exit 1
    fi
}

# Run checkpoint
checkpoint_run() {
    cmd=$*
    if [ -f $CHECKLOG ] && grep -q "$cmd" "$CHECKLOG"; then
        log "Retrying checkpoint: $cmd"
        cmd_exe $cmd
        rm "$CHECKLOG"
    else
        cmd_exe $cmd
    fi
}

# Manage Go modules
go_mod() {
    go mod download all
    go mod tidy
}

# Discard Go module changes
go_discard() {
    git checkout go.mod go.sum
}

# Main script
main() {
    process_arguments "$@"
    restart_upgrade
    initialize_upgrade
    start_validation
    setup_branches
    merge_branches
    generate_patch
}

# --- start ---
CHECKLOG=$WORKSPACE/pbs-patch/checkpoints.log
trap 'clear_log' EXIT

log "Attempt: $attempt"
main "$@"
