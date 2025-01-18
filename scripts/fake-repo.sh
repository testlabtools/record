#!/usr/bin/env bash
set -eu -o pipefail

repo="$1"
clone="$2"

rm -rf "$repo" "$clone"

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

mkdir -p "$repo/.github" "$clone"

cat -> "$repo/.github/CODEOWNERS" <<EOF
*.go @org/team1
e2e/ @org/team2
EOF

# Add codeowners
add "add owners" .github/CODEOWNERS

# Clone repo to have a remote.
git clone --template=/dev/null "$repo" "$clone"

# Set main as default remote origin branch.
git -C "$repo" remote add origin "$(realpath $clone)"
git -C "$repo" fetch origin
git -C "$repo" branch -u origin/main
