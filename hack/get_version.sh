#!/usr/bin/env bash

set -e

: "${GIT_CMD:=git}"

function _git {
    ${GIT_CMD} "$@"
}

function git_describe {
    _git describe --tags --first-parent "$@"
}

function get_last_tag {
    # If no tags yet, use 0.0.0 as base
    git_describe --abbrev=0 2>/dev/null || echo "0.0.0"
}

function trim_tag {
    echo "${1#v}"
}

function increment_version {
    IFS=. read -r -a parts <<< "$1"
    if [ "${#parts[@]}" -ne 3 ]; then
        echo "Bad version: ${1}" 1>&2
        exit 1
    fi
    echo "${parts[0]}.${parts[1]}.$((parts[2]+1))"
}

function main {
    last_tag="$(get_last_tag)"
    trimmed_tag="$(trim_tag "${last_tag}")"
    next_version="$(increment_version "${trimmed_tag}")"
    commit_sha="$(_git rev-parse --short=7 HEAD)"

    # If current commit is tagged, version is whatever tag says
    if git_describe --exact-match &> /dev/null; then
        echo "${trimmed_tag}"
        return
    fi

    if [ "${last_tag}" = "0.0.0" ]; then
        commits_since_tag="$(_git rev-list --count HEAD)"
    else
        commits_since_tag="$(_git rev-list --count "${last_tag}"..HEAD)"
    fi

    echo "${next_version}-${commits_since_tag}-${commit_sha}"
}

main "$@"
