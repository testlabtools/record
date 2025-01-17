#!/usr/bin/env bash
set -eu -o pipefail

repo="$1"

rm -rf "$repo"

# dont copy template hook files
git init "$repo" --template=/dev/null

function set_user() {
    local name="$1"
    local email="$2"
    git -C "$repo" config --local user.name "$name"
    git -C "$repo" config --local user.email "$email"
}

function add() {
    local msg="$1"
    shift

    git -C "$repo" add $@
    git -C "$repo" commit -m "$msg"
}

set_user "User One" "user1@org"

mkdir "$repo/.github"

cat -> "$repo/.github/CODEOWNERS" <<EOF
*.go @org/team1
e2e/ @org/team2
EOF

# Add codeowners
add "add owners" .github/CODEOWNERS
