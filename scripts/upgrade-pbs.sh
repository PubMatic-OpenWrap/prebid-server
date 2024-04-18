#!/bin/bash -e


attempt=1

usage="
Script starts or continues prebid upgrade to version set in 'to_minor' variable. Workspace is at /tmp/prebid-server and /tmp/pbs-patch

    ./upgrade-pbs.sh [--restart]

    --restart   Restart the upgrade (deletes /tmp/prebid-server and /tmp/pbs-patch)
    -h          Help

TODO:
    - paramertrize the script
    - create ci branch PR
    - create header-bidding PR"

RESTART=0
for i in "$@"; do
  case $i in
    --restart)
      RESTART=1
      shift
      ;;
    -h)
      echo "$usage"
      exit 0
      ;;
  esac
done

# --- start ---
CHECKLOG=/tmp/pbs-patch/checkpoints.log

trap 'clear_log' EXIT

log () {
  printf "\n$(date): $1\n"
}

clear_log() {
    current_fork_at_version=$(git describe --tags --abbrev=0)
    if [ "$current_fork_at_version" == "$upgrade_version" ] ; then
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


clone_repo() {
    if [ -d "/tmp/prebid-server" ]; then
        log "Code already cloned. Attempting to continue the upgrade!!!"
    else
        log "Cloning repo at /tmp"
        cd /tmp
        git clone https://github.com/PubMatic-OpenWrap/prebid-server.git
        cd prebid-server

        git remote add prebid-upstream https://github.com/prebid/prebid-server.git
        git remote -v
        git fetch --all --tags --prune
    fi
}

checkout_branch() {
    set +e
    git checkout tags/$_upgrade_version -b $tag_base_branch_name
    # git push origin $tag_base_branch_name

    git checkout -b $upgrade_branch_name
    git checkout $upgrade_branch_name
    # git push origin $upgrade_branch_name

    set -e
#    if [ "$?" -ne 0 ]
#    then
#        log "Failed to create branch $upgrade_branch_name. Already working on it???"
#        exit 1
#    fi
}

cmd_exe() {
    cmd=$*
    if ! $cmd; then
        log "Failure!!! creating checkpoint $cmd"
        echo "$cmd" > $CHECKLOG
        exit 1
    fi
}

checkpoint_run() {
    cmd=$*
    if [ -f $CHECKLOG ] ; then
        if grep -q "$cmd" "$CHECKLOG"; then
            log "Retry this checkpoint: $cmd"
            cmd_exe $cmd
            rm "$CHECKLOG"
        elif grep -q "./validate.sh --race 5 --nofmt" "$CHECKLOG"; then
            log "Special checkpoint. ./validate.sh --race 5 --nofmt failed for last tag update. Hence, only fixes are expected in successfully upgraded branch. (change in func() def, wrong conflict resolve, etc)"
            cmd_exe $cmd
            rm "$CHECKLOG"
        else
            log "Skip this checkpoint: $cmd"
            return
        fi
    fi
    cmd_exe $cmd
}

go_mod() {
    go mod download all
    go mod tidy
    go mod tidy
    go mod download all
}

go_discard() {
    git checkout go.mod go.sum
}

# --- main ---

if [ "$RESTART" -eq "1" ]; then
    log "Restarting the upgrade: rm -rf /tmp/prebid-server /tmp/pbs-patch/"
    rm -rf /tmp/prebid-server /tmp/pbs-patch/ 
    mkdir -p /tmp/pbs-patch/
fi

log "Final Upgrade Version: $upgrade_version"
log "Attempt: $attempt"

checkpoint_run clone_repo
cd /tmp/prebid-server
log "At $(pwd)"

# Get the latest tag
latest_tag=$(git describe --tags --abbrev=0)

git diff tags/$latest_tag..origin/master > /tmp/pbs-patch/current_ow_patch-$latest_tag-origin_master-$attempt.diff

log "Starting with version :$latest_tag"

log "Checking if last failure was for test case. Need this to pick correct"
go_mod
checkpoint_run "./validate.sh --race 5 --nofmt"
go_discard

# Loop through each tag and merge it
tags=$(git tag --merged prebid-upstream/master --sort=v:refname)

log "Starting upgrade loop..."
for tag in $tags
    do
    if [[ "$tag" == "$latest_tag" ]]; then
        found_latest_tag=true
        if [[ -f $CHECKLOG ]]; then
            log "At tag: $tag but $CHECKLOG exists. Continue last failed checkpoint."
        else
            continue
        fi
    fi

    if [[ "$found_latest_tag" = true  && "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    _upgrade_version=$tag

    log "Starting upgrade to version $_upgrade_version"

    tag_base_branch_name=prebid_$_upgrade_version-$attempt-tag
    upgrade_branch_name=prebid_$_upgrade_version-$attempt
    log "Reference tag branch: $tag_base_branch_name"
    log "Upgrade branch: $upgrade_branch_name"

    checkpoint_run checkout_branch

    log "Merging master in $tag_base_branch_name"
    checkpoint_run git merge master --no-edit
    # Use `git commit --amend --no-edit` if you had to fix test cases, etc for wrong merge conflict resolve, etc.
    log "Validating the master merge into current tag. Fix and commit changes if required. Use 'git commit --amend --no-edit' for consistency"
    go_mod
    checkpoint_run "./validate.sh --race 5 --nofmt"
    go_discard

    checkpoint_run git checkout master
    checkpoint_run git merge $upgrade_branch_name --no-edit

    log "Generating patch file at /tmp/pbs-patch/ for $_upgrade_version"
    git diff tags/$_upgrade_version..master > /tmp/pbs-patch/new_ow_patch_$upgrade_version-master-1.diff
    fi
done
