#!/bin/bash -e

prefix="v"
to_major=0
to_minor=205
to_patch=0

attempt=3

log () {
  printf "\n$(date): $1\n"
}

get_current_tag_version() {
    log "get_current_tag_version $@"

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

upgrade_version="$prefix$to_major.$to_minor.$to_patch"
log "Final Upgrade Version: $upgrade_version"
log "Attempt: $attempt"

if [ -d "/tmp/prebid-server" ]; then
    log "Code already cloned. Attempting to continue the upgrade!!!"
    cd /tmp/prebid-server
else
    cd /tmp
    git clone https://github.com/PubMatic-OpenWrap/prebid-server.git
    cd prebid-server

    git remote add prebid-upstream https://github.com/prebid/prebid-server.git
    git remote -v
    git fetch --all --tags --prune
fi

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

    set +e
    git checkout tags/$_upgrade_version -b $tag_base_branch_name
    # git push origin $tag_base_branch_name
    set -e

    git checkout -b $upgrade_branch_name
    # git push origin $upgrade_branch_name

    if [ "$?" -ne 0 ]
    then
        log "Failed to create branch $upgrade_branch_name. Already working on it???"
        exit 2
    fi

    git merge master --no-edit
    git checkout master
    git merge $upgrade_branch_name --no-edit

    git diff tags/$_upgrade_version..master > new_ow_patch_$upgrade_version-master-1.diff

    log "Validating the merge for current tag"
    ./validate.sh
done

# TODO:
# Open draft PR between $tag_base_branch_name <- $upgrade_branch_name
# What changes do we have on top of prebid's tag

# git rebase master

#todo check why code fails even with correct rebase

# for each conflict
#     1. resolve conflict
#     2. move OpenWrap exclusive code to new _pubmatic.go file to avoid future conflicts -> TODO: DONT DO THIS HERE, BREAKS COMMIT HISTORY
#     3. go mod tidy
#     4. git commit -m "<SAME COMMIT MESSAGE AS OF CONFLICT. DO NOT CHANGE THIS MSG>"
#     5. git rebase --continue

# go mod tidy
# go mod download all
# go mod tidy
# go mod download all
# ./validate.sh

# This commit should be of tag $upgrade_version
# git merge-base prebid-upstream/master origin/$upgrade_branch_name
# After master merge, this commit should be of tag $upgrade_version
# git merge-base prebid-upstream/master origin/master

