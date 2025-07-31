#!/bin/bash

# Exit on any error
set -e

# Default values
VERBOSE="FALSE"
CHANGELOG_FILE="CHANGELOG.md"
RELEASE_NOTES_FILE="release_notes.md"
OUTPUT_TO_FILE="FALSE"
MODE="GENERATE"
REPO_NAME=""
NEW_RELEASE=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
  -m)
    MODE=$2
    shift 2
    ;;
  --mode)
    MODE=$2
    shift 2
    ;;
  -v)
    NEW_RELEASE=$2
    shift 2
    ;;
  --version)
    NEW_RELEASE=$2
    shift 2
    ;;
  -r)
    REPO_NAME=$2
    shift 2
    ;;
  --repo)
    REPO_NAME=$2
    shift 2
    ;;
  --CHANGELOG_FILE)
    CHANGELOG_FILE=$2
    shift 2
    ;;
  --file)
    RELEASE_NOTES_FILE=$2
    shift 2
    ;;
  --output-to-file)
    OUTPUT_TO_FILE="TRUE"
    shift
    ;;
  --verbose)
    VERBOSE="TRUE"
    shift
    ;;
  *)
    echo -e "${RED}Error: Invalid argument: $1${NC}" >&2
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -m, --mode MODE              Mode: GENERATE or RELEASE"
    echo "  -v, --version VERSION        Version number for the release"
    echo "  -r, --repo REPO              Repository name (owner/repo)"
    echo "  --CHANGELOG_FILE FILE        Changelog file path (default: CHANGELOG.md)"
    echo "  --file FILE                  Release notes file path (default: release_notes.md)"
    echo "  --output-to-file             Output to file instead of stdout"
    echo "  --verbose                    Enable verbose output"
    exit 1
    ;;
  esac
done

# Validation function
validate_requirements() {
  # Check if gh CLI is available
  if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed or not in PATH${NC}" >&2
    exit 1
  fi

  # Check if jq is available
  if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is not installed or not in PATH${NC}" >&2
    exit 1
  fi

  # Validate required parameters
  if [ -z "$REPO_NAME" ]; then
    echo -e "${RED}Error: Repository name is required (-r or --repo)${NC}" >&2
    exit 1
  fi

  if [ "$MODE" == "GENERATE" ] && [ -z "$NEW_RELEASE" ]; then
    echo -e "${RED}Error: Version is required for GENERATE mode (-v or --version)${NC}" >&2
    exit 1
  fi

  # Validate repository format
  if [[ ! "$REPO_NAME" =~ ^[^/]+/[^/]+$ ]]; then
    echo -e "${RED}Error: Repository name must be in format 'owner/repo'${NC}" >&2
    exit 1
  fi
}

# Log function
log() {
  if [ "$VERBOSE" == "TRUE" ]; then
    echo -e "${GREEN}[INFO]${NC} $1"
  fi
}

# Log function without colors (for file output)
log_no_color() {
  if [ "$VERBOSE" == "TRUE" ]; then
    echo "[INFO] $1"
  fi
}

log_warn() {
  echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Extract changelog content from PR body
extract_changelog_content() {
  local pr_body="$1"
  local temp_file=$(mktemp)
  
  # Extract content from # Description to next ##
  echo "$pr_body" | awk '
    /^# Description$/ { in_description = 1; next }
    /^##/ { in_description = 0; next }
    in_description { print }
  ' > "$temp_file"

  # Extract only lines starting with -
  local changelog_lines=""
  while IFS= read -r line; do
    # Skip empty lines and whitespace-only lines
    if [[ -n "$line" && ! "$line" =~ ^[[:space:]]*$ ]]; then
      # Check if line starts with - (after any leading whitespace)
      if [[ "$line" =~ ^[[:space:]]*- ]]; then
        # Clean up the line and add to changelog
        cleaned_line=$(echo "$line" | sed 's/^[[:space:]]*//')
        if [ -n "$changelog_lines" ]; then
          changelog_lines="$changelog_lines"$'\n'"$cleaned_line"
        else
          changelog_lines="$cleaned_line"
        fi
      fi
    fi
  done < "$temp_file"

  rm "$temp_file"
  echo "$changelog_lines"
}

# Get the last release date
get_last_release_date() {
  local last_release_date=""
  
  # Try to get the last PR with release-request label
  local last_release_pr=$(gh pr list --repo "$REPO_NAME" --base main --json mergedAt --state merged --search "label:release-request" --limit 1 2>/dev/null || true)
  
  if [ -n "$last_release_pr" ] && [ "$last_release_pr" != "[]" ]; then
    last_release_date=$(echo "$last_release_pr" | jq -r '.[0].mergedAt // empty' 2>/dev/null || true)
  fi
  
  echo "$last_release_date"
}

# Generate release notes from PRs
generate_release_notes() {
  local last_release_date="$1"
  local temp_file=$(mktemp)
  local all_changes=""
  
  # Only log if we're not being called for content generation
  if [ "$OUTPUT_TO_FILE" != "TRUE" ]; then
    log_no_color "Generating release notes for repository $REPO_NAME"
    
    # Build the search query
    local search_query="--base main --state merged"
    if [ -n "$last_release_date" ]; then
      search_query="$search_query --search \"merged:>$last_release_date\""
      log_no_color "Looking for PRs merged after $last_release_date"
    else
      log_no_color "No previous release found, including all merged PRs"
    fi
  else
    # Build the search query without logging
    local search_query="--base main --state merged"
    if [ -n "$last_release_date" ]; then
      search_query="$search_query --search \"merged:>$last_release_date\""
    fi
  fi
  
  # Get all relevant PRs
  local prs=$(gh pr list --repo "$REPO_NAME" $search_query --json body,title,number 2>/dev/null || true)
  
  if [ -z "$prs" ] || [ "$prs" == "[]" ]; then
    if [ "$OUTPUT_TO_FILE" != "TRUE" ]; then
      log_no_color "No PRs found for this release"
    fi
    echo "" > "$temp_file"
  else
    # Process each PR
    echo "$prs" | jq -r '.[] | @base64' | while read -r pr_base64; do
      if [ -n "$pr_base64" ]; then
        local pr_data=$(echo "$pr_base64" | base64 --decode)
        local pr_body=$(echo "$pr_data" | jq -r '.body // ""')
        local pr_title=$(echo "$pr_data" | jq -r '.title // ""')
        local pr_number=$(echo "$pr_data" | jq -r '.number // ""')
        
        # Extract changelog content
        local changelog_content=$(extract_changelog_content "$pr_body")
        
        if [ -n "$changelog_content" ]; then
          if [ "$OUTPUT_TO_FILE" != "TRUE" ]; then
            log_no_color "Found changelog content in PR #$pr_number: $pr_title"
          fi
          if [ -n "$all_changes" ]; then
            all_changes="$all_changes"$'\n'"$changelog_content"
          else
            all_changes="$changelog_content"
          fi
        else
          if [ "$OUTPUT_TO_FILE" != "TRUE" ]; then
            log_no_color "No changelog content found in PR #$pr_number: $pr_title"
          fi
        fi
      fi
    done
    
    # Remove duplicates and sort
    if [ -n "$all_changes" ]; then
      echo "$all_changes" | sort -u > "$temp_file"
    else
      echo "" > "$temp_file"
    fi
  fi
  
  # Read the final content
  local content=$(cat "$temp_file")
  rm "$temp_file"
  
  # Output based on mode
  if [ "$OUTPUT_TO_FILE" == "TRUE" ]; then
    if [ -f "$RELEASE_NOTES_FILE" ]; then
      rm "$RELEASE_NOTES_FILE"
    fi
    echo "# Release $NEW_RELEASE" > "$RELEASE_NOTES_FILE"
    echo "" >> "$RELEASE_NOTES_FILE"
    echo "$content" >> "$RELEASE_NOTES_FILE"
    if [ "$OUTPUT_TO_FILE" != "TRUE" ]; then
      log "Release notes saved to $RELEASE_NOTES_FILE"
    fi
  else
    echo "$content"
  fi
}

# Insert content into changelog file
insert_changelog_content() {
  local line_number="$1"
  local content="$2"
  local file="$3"
  
  local temp_file=$(mktemp)
  local content_file=$(mktemp)
  
  echo "$content" > "$content_file"
  
  awk -v lineno="$line_number" -v content_file="$content_file" '
    NR == lineno {
        while ((getline line < content_file) > 0) {
            print line
        }
        close(content_file)
    }
    { print }
  ' "$file" > "$temp_file"
  
  mv "$temp_file" "$file"
  rm "$content_file"
}

# Append content to existing changelog section
append_changelog_content() {
  local start_line="$1"
  local end_line="$2"
  local content="$3"
  local file="$4"
  
  local temp_file=$(mktemp)
  local content_file=$(mktemp)
  
  echo "$content" > "$content_file"
  end_line=$((end_line - 1))
  
  awk -v start="$start_line" -v end="$end_line" -v content_file="$content_file" '
    {
        print
        if (NR == end) {
            while ((getline line < content_file) > 0) {
                print line
            }
            close(content_file)
        }
    }
  ' "$file" > "$temp_file"
  
  mv "$temp_file" "$file"
  rm "$content_file"
}

# Generate changelog entry
generate_changelog_entry() {
  local temp_file=$(mktemp)
  local content_file=$(mktemp)
  
  log "Generating changelog entry for version $NEW_RELEASE"
  
  # Get last release date
  local last_release_date=$(get_last_release_date)
  
  # Generate release notes to a separate file to avoid log contamination
  OUTPUT_TO_FILE=TRUE generate_release_notes "$last_release_date" > "$content_file" 2>/dev/null
  local content=$(cat "$content_file")
  
  # Check if version already exists
  local version_line=$(grep -n "^## \[$NEW_RELEASE\]" "$CHANGELOG_FILE" 2>/dev/null | cut -d: -f1 || true)
  
  if [ -z "$version_line" ]; then
    log "Version $NEW_RELEASE does not exist, creating new section"
    
    # Find where to insert the new version
    local header_end_line=$(grep -n -m1 "^## \[.*\]" "$CHANGELOG_FILE" 2>/dev/null | cut -d: -f1 || true)
    local insert_line=3
    
    if [ -z "$header_end_line" ]; then
      # No existing versions, append at the end
      insert_line=$(wc -l < "$CHANGELOG_FILE")
      insert_line=$((insert_line + 1))
    else
      # Insert before the first existing version
      insert_line=$((header_end_line - 1))
    fi
    
    local today=$(date '+%Y-%m-%d')
    local new_version_section="## [$NEW_RELEASE] - $today

"
    
    # Only add content if there is any
    if [ -n "$content" ]; then
      new_version_section="${new_version_section}${content}

"
    fi
    
    insert_changelog_content "$insert_line" "$new_version_section" "$CHANGELOG_FILE"
  else
    log "Version $NEW_RELEASE exists, appending content"
    
    # Find where the version section ends
    local next_version_line=$(awk -v ver_line="$version_line" 'NR > ver_line && /^## \[.*\]/ {print NR; exit}' "$CHANGELOG_FILE" 2>/dev/null || true)
    
    local end_line
    if [ -z "$next_version_line" ]; then
      # Version section goes to the end of the file
      end_line=$(wc -l < "$CHANGELOG_FILE")
    else
      end_line=$((next_version_line - 1))
    fi
    
    append_changelog_content "$version_line" "$end_line" "$content" "$CHANGELOG_FILE"
  fi
  
  rm "$temp_file"
  log "Changelog updated successfully"
}

# Main execution
main() {
  # Validate requirements
  validate_requirements
  
  # Check if changelog file exists, create if not
  if [ ! -f "$CHANGELOG_FILE" ]; then
    log "Creating new changelog file: $CHANGELOG_FILE"
    cat > "$CHANGELOG_FILE" << EOF
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

EOF
  fi
  
  case "$MODE" in
    "GENERATE")
      generate_changelog_entry
      ;;
    "RELEASE")
      local last_release_date=$(get_last_release_date)
      generate_release_notes "$last_release_date"
      ;;
    *)
      log_error "Invalid mode: $MODE"
      exit 1
      ;;
  esac
}

# Run main function
main "$@"