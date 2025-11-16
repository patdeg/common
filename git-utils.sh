#!/usr/bin/env bash
# --------------------------------------------
# Git Utilities — safer, faster day-to-day Git
# Usage:
#   source /path/to/git-utils.sh
#   ghelp
# --------------------------------------------

# Fail early when not inside a git repo
_g_require_repo() {
  git rev-parse --git-dir >/dev/null 2>&1 || {
    echo "Not a git repository (or any of the parent directories)."; return 1;
  }
}

# Show current branch name
_g_branch() {
  git rev-parse --abbrev-ref HEAD 2>/dev/null
}

# Timestamp helper
_g_stamp() { date +%Y%m%d-%H%M%S; }

# -------------------------------
# Commit workflow shortcuts
# -------------------------------

# Stage everything (git add -A)
gstage() {
  _g_require_repo || return 1
  git add -A
  echo "✅ Staged all changes."
}

# Commit with a message
gcommit() {
  _g_require_repo || return 1
  if [ -z "$1" ]; then
    echo "Usage: gcommit \"your commit message\""; return 1
  fi
  git commit -m "$1" && echo "✅ Committed: $1"
}

# Push current branch
gpush() {
  _g_require_repo || return 1
  local br=$(_g_branch)
  git push origin "$br" && echo "⬆️  Pushed $br to origin."
}

# Push current branch and set upstream if needed
gpushu() {
  _g_require_repo || return 1
  local br=$(_g_branch)
  git push -u origin "$br" && echo "⬆️  Pushed $br (upstream set)."
}

# -------------------------------
# Branch / Tag safety
# -------------------------------

# Checkout a tag into a new branch
gco_tag() {
  _g_require_repo || return 1
  if [ -z "$1" ]; then
    echo "Usage: gco_tag <tagname> [branchname]"; return 1
  fi
  local tag="$1"
  local branch="${2:-from-$tag}"
  git checkout -b "$branch" "$tag" && echo "✅ Checked out tag $tag into branch $branch"
}

# Backup current HEAD with a timestamped tag
gbackup() {
  _g_require_repo || return 1
  local tag="backup-$(_g_stamp)"
  git tag "$tag" && echo "🧩 Tagged HEAD as $tag"
}

# Reset current branch to a tag (HARD reset)
greset_tag() {
  _g_require_repo || return 1
  if [ -z "$1" ]; then
    echo "Usage: greset_tag <tagname>"; return 1
  fi
  git reset --hard "$1" && echo "🧹 Branch reset to tag $1"
}

# Force-sync current branch with remote (safe variant)
gforce_sync() {
  _g_require_repo || return 1
  local br=$(_g_branch)
  git push --force-with-lease origin "$br" && echo "🔁 Force-synced $br with --force-with-lease"
}

# One-button “panic” backup (tag + branch)
gpanic() {
  _g_require_repo || return 1
  local stamp=$(_g_stamp)
  local br=$(_g_branch)
  local tag="panic-$stamp"
  local safebr="safety-$br-$stamp"
  git tag "$tag" && echo "🆘 Panic tag: $tag"
  git checkout -b "$safebr" && echo "🆘 Safety branch: $safebr (at $(git rev-parse --short HEAD))"
}

# -------------------------------
# Info & sync
# -------------------------------

# Show branches sorted by latest commit
gbranches() {
  _g_require_repo || return 1
  git branch --sort=-committerdate
}

# List tags sorted by creation date
gtags() {
  _g_require_repo || return 1
  git for-each-ref --sort=-creatordate --format '%(refname:short)  %(creatordate:short)' refs/tags
}

# Safe pull with rebase
gpull() {
  _g_require_repo || return 1
  git pull --rebase --autostash && echo "⬇️  Rebased with autostash."
}

# Status (short & clean)
gstatus() {
  _g_require_repo || return 1
  git status -sb
}

# Pretty log
glog() {
  _g_require_repo || return 1
  git log --oneline --decorate --graph --max-count="${1:-20}"
}

# Diff summary
gdiff() {
  _g_require_repo || return 1
  git diff --stat "${1:-HEAD}"
}

# -------------------------------
# Help
# -------------------------------
ghelp() {
  cat <<'EOF'
Git Utilities — safer shortcuts

Commit workflow:
  gstage                   Stage all changes (git add -A)
  gcommit "msg"            Commit staged changes
  gpush                    Push current branch
  gpushu                   Push and set upstream

Branch / Tag safety:
  gco_tag <tag> [branch]   Checkout a tag into a new branch
  gbackup                  Tag current HEAD with timestamp
  greset_tag <tag>         HARD reset current branch to tag
  gforce_sync              Force-push current branch with lease
  gpanic                   Create panic tag + safety branch

Info & sync:
  gstatus                  Compact status (git status -sb)
  gbranches                Branches by last commit time
  gtags                    Tags by creation date
  gpull                    Pull with rebase + autostash
  glog [N]                 Pretty graph log (default 20)
  gdiff [ref]              Diff summary vs ref (default HEAD)

Usage:
  source /path/to/git-utils.sh
  ghelp

Safety tips:
  • Run gbackup before risky operations (rebase/reset).
  • Prefer gforce_sync (with lease) over raw --force.
  • Use gco_tag to explore old states safely on a new branch.
EOF
}


