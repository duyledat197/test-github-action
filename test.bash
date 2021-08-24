set -e
last_release=$(git branch --all --list '*release/20*' | sort -r | head -n 1 | cut -c18-)
if [[ $GITHUB_BASE_REF == $last_release ]];
then
    echo "ERROR: target branch is old (got \"${GITHUB_BASE_REF}\", expected \"${last_release}\")"; return 1
fi
return 0