#!/bin/bash -e

prefix="v"
to_major=0
to_minor=205
to_patch=0
upgrade_version="$prefix$to_major.$to_minor.$to_patch"

attempt=3

CHECKLOG=checkpoints.log

trap 'clear_log' EXIT

log () {
  printf "\n$(date): $1\n"
}

clear_log() {
    major=0
    minor=0
    patch=0
    get_current_tag_version major minor patch
    current_fork_at_version="$major.$minor.$patch"

    if [ "$current_fork_at_version" == "$upgrade_version" ] ; then
        log "Upgraded to $current_fork_at_version"
            rm -f "$CHECKLOG"
    else
        log "Exiting with failure!!!"
    fi
}

get_current_tag_version() {
    log "get_current_tag_version $*"

    local -n _major=$1
    local -n _minor=$2
    local -n _patch=$3

    # script will always start from start if origin/master is used.
    # common_commit=$(git merge-base prebid-upstream/master origin/master)
    # log "Common commit b/w prebid-upstream/master origin/master: $common_commit"

    # remove origin for master to continue from last fixed tag's rebase.
    common_commit=$(git merge-base prebid-upstream/master master)
    log "Common commit b/w prebid-upstream/master master: $common_commit"

    current_version=$(git tag --points-at $common_commit)
    if [[ $current_version == v* ]] ; then
        log "Current Version: $current_version"
    else
        log "Failed to detected current version. Abort."
        exit 1
        # abort
        # cd prebid-server; git rebase --abort;cd -
    fi

    IFS='.' read -r -a _current_version <<< "$current_version"
    _major=${_current_version[0]}
    _minor=${_current_version[1]}
    _patch=${_current_version[2]}
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
    set -e

    git checkout -b $upgrade_branch_name
    # git push origin $upgrade_branch_name

    if [ "$?" -ne 0 ]
    then
        log "Failed to create branch $upgrade_branch_name. Already working on it???"
        exit 1
    fi
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
            rm "$CHECKLOG"
        elif grep -q "./validate.sh" "$CHECKLOG"; then
            log "Special checkpoint. ./validate.sh failed for last tag update. Hence, only fixes are expected in successfully upgraded branch. (change in func() def, wrong conflict resolve, etc)"
            cmd_exe $cmd
            rm "$CHECKLOG"
        else
            log "Skip this checkpoint: $cmd"
            return
        fi
    fi
    cmd_exe $cmd
}

# --- main ---

log "Final Upgrade Version: $upgrade_version"
log "Attempt: $attempt"

checkpoint_run clone_repo
cd /tmp/prebid-server
log "At $(pwd)"

major=0
minor=0
patch=0

get_current_tag_version major minor patch
((minor++))
log "Starting with version split major:$major, minor:$minor, patch:$patch"
current_fork_at_version="$major.$minor.$patch"
git diff tags/$current_fork_at_version..origin/master > current_ow_patch-$current_fork_at_version-origin_master-$attempt.diff

while [ "$minor" -le "$to_minor" ]; do
    # _upgrade_version="$prefix$major.$minor.$patch"
    _upgrade_version="$major.$minor.$patch"
    ((minor++))

    log "Starting upgrade to version $_upgrade_version"

    tag_base_branch_name=prebid_$_upgrade_version-$attempt-tag
    upgrade_branch_name=prebid_$_upgrade_version-$attempt
    log "Reference tag branch: $tag_base_branch_name"
    log "Upgrade branch: $upgrade_branch_name"

    checkpoint_run checkout_branch

    checkpoint_run git merge master --no-edit
    checkpoint_run git checkout master
    checkpoint_run git merge $upgrade_branch_name --no-edit

    git diff tags/$_upgrade_version..master > new_ow_patch_$upgrade_version-master-1.diff

    # Use `git commit --amend --no-edit` if you had to fix test cases, etc for wrong merge conflict resolve, etc.
    log "Validating the merge for current tag"
    go mod download all
    go mod tidy
    go mod tidy
    go mod download all
    checkpoint_run "./validate.sh"

    # revert changes by ./validate.sh
    git checkout master go.mod
    git checkout master go.sum
done