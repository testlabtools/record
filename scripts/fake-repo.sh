#!/usr/bin/env bash
set -eu -o pipefail

repo="$1"
feature="$2"

rm -rf "$repo" "$feature"

# dont copy template hook files
git init "$repo" --template=/dev/null
git -C "$repo" branch -m main

function set_user() {
    local name="$1"
    local email="$2"
    git -C "$repo" config --local user.name "$name"
    git -C "$repo" config --local user.email "$email"
}

function add() {
    local repo="$1"
    local msg="$2"
    shift
    shift

    git -C "$repo" add $@
    git -C "$repo" commit -m "$msg"
}

set_user "User One" "user1@org"

mkdir -p "$repo/.github" "$feature"

cat -> "$repo/.github/CODEOWNERS" <<EOF
*.go @org/team1
e2e/ @org/team2
EOF

# Add codeowners
add "$repo" "add owners" .github/CODEOWNERS
git -C "$repo" tag 1.0.2

# Clone repo to have a remote.
git clone --template=/dev/null "$repo" "$feature"

# Set main as default remote origin branch.
git -C "$repo" remote add origin "$(realpath $feature)"
git -C "$repo" fetch origin
git -C "$repo" branch -u origin/main

# Create feature branch and tag.
git -C "$feature" checkout -b my-feature
git -C "$feature" tag 2.my-feature.3
