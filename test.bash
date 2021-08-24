set -e
last_release=$(git branch --all --list '*release/20*' | sort -r | head -n 1 | cut -c18-)
GITHUB_BASE_REF=1
echo $(git branch --all --list '*release/20*' | sort -r | head -n 1 | cut -c18-)
echo ${GITHUB_BASE_REF}
echo $last_release
if [[ $GITHUB_BASE_REF != $last_release ]];
then
    echo "ERROR: target branch is old (got \"${GITHUB_BASE_REF}\", expected \"${last_release}\")"; 
    return 1
fi