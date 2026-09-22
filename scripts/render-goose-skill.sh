#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
    echo "usage: $0 <source> <installed-name>" >&2
    exit 2
fi

source_file=$1
installed_name=$2

awk -v installed_name="$installed_name" '
BEGIN { delimiters = 0; description = "" }
$0 == "---" {
    delimiters++
    if (delimiters == 2) {
        if (description == "") {
            print "missing description in " FILENAME > "/dev/stderr"
            exit 1
        }
        print "---"
        print "name: " installed_name
        print description
        print "---"
    }
    next
}
delimiters == 1 && /^description:[[:space:]]*/ { description = $0; next }
delimiters >= 2 { print }
END {
    if (delimiters < 2) {
        print "invalid frontmatter in " FILENAME > "/dev/stderr"
        exit 1
    }
}
' "$source_file"
