set -e
check_format_release_branch() {
            regexp_release_branch="^release\/(20[2-9][1-9])(1[0-2]|0[1-9])(3[01]|[0-2][1-9]|[12]0)$"
            branch="$1"
            match_count=$(echo ${branch} | grep -o -P "${regexp_release_branch}" | wc -l)
            if [[ ${match_count} != 1 ]]; then
              >&2 echo "ERROR: release branch is not properly formatted (got \"${branch}\", expected \"release/yyyymmdd\")"; return 1
            fi
            return 0
          }
         
last_release=$(git branch --all --list '*release/20*' | sort -r | head -n 1 | cut -c18-)
GITHUB_BASE_REF=1
echo $(git branch --all --list '*release/20*' | sort -r | head -n 1 | cut -c18-)
echo ${GITHUB_BASE_REF}
echo $last_release
if [[  $(check_format_release_branch "${BRANCH}") && $GITHUB_BASE_REF != $last_release ]];
then
    echo "ERROR: target branch is old (got \"${GITHUB_BASE_REF}\", expected \"${last_release}\")"; 
    return 1
fi