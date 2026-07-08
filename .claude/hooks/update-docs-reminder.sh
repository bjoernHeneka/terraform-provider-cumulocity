#!/usr/bin/env bash
# Reminds Claude to keep documentation in sync whenever a provider
# resource or data source Go file is written or edited.
#
# Triggered as a PostToolUse hook for Write and Edit tools.
# Receives the tool JSON payload on stdin; outputs a reminder on stdout
# which Claude Code injects back into the conversation context.

set -euo pipefail

payload=$(cat)

# Extract tool_name and file_path from the JSON payload.
tool_name=$(echo "$payload" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_name',''))" 2>/dev/null || true)
file_path=$(echo "$payload" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('file_path',''))" 2>/dev/null || true)

[[ -z "$file_path" ]] && exit 0

# Only act on resource or data source Go files inside internal/provider/.
if [[ ! "$file_path" =~ internal/provider/(.+)_(resource|data_source)\.go$ ]]; then
    exit 0
fi

base="${BASH_REMATCH[1]}"
kind="${BASH_REMATCH[2]}"

if [[ "$kind" == "resource" ]]; then
    doc_file="docs/resources/${base}.md"
    doc_label="resource"
else
    doc_file="docs/data-sources/${base}.md"
    doc_label="data source"
fi

# Resolve project root from the file path so the doc existence check works
# regardless of the shell's current working directory.
project_root=$(echo "$file_path" | sed 's|/internal/provider/.*||')
doc_abs="${project_root}/${doc_file}"

if [[ "$tool_name" == "Write" && ! -f "$doc_abs" ]]; then
    cat <<MSG
DOCUMENTATION REMINDER: A new ${doc_label} was created at '${file_path}'.
Create the matching documentation file '${doc_file}' now, following the existing docs structure:
  - Frontmatter: page_title and description
  - Brief description paragraph
  - ## Example Usage  (with HCL code block)
  - ## Schema  (Required / Optional / Read-Only subsections)
  - ## Import  (terraform import shell example)
MSG
else
    cat <<MSG
DOCUMENTATION REMINDER: '${file_path}' was modified.
Review '${doc_file}' and update it if the schema, attribute descriptions, import format, or behavior changed.
MSG
fi

exit 0
