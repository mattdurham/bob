#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
    echo "usage: $0 <source> <companion-directory>" >&2
    exit 2
fi

source_file=$1
companion_dir=$2
case "$companion_dir" in
    *'|'*|*'&'*|*'\'*|*'"'*|*$'\n'*)
        echo "unsafe companion directory: $companion_dir" >&2
        exit 1
        ;;
esac

awk '
BEGIN { delimiters = 0; name = ""; description = "" }
$0 == "---" {
    delimiters++
    if (delimiters == 2) {
        if (name == "" || description == "") {
            print "missing name or description in " FILENAME > "/dev/stderr"
            exit 1
        }
        print "---"
        print name
        print description
        print "---"
    }
    next
}
delimiters == 1 && /^name:[[:space:]]*/ { name = $0; next }
delimiters == 1 && /^description:[[:space:]]*/ { description = $0; next }
delimiters >= 2 { print }
END {
    if (delimiters < 2) {
        print "invalid frontmatter in " FILENAME > "/dev/stderr"
        exit 1
    }
}
' "$source_file" | sed "s|\[agent-directory\]|$companion_dir|g"
