async function getCommitOnPR({ github, context }, prNumber) {
  let listCommitInfo = [];
  const limit = 100;
  let page = 1;
  let dataSize = 0;

  console.log(prNumber);
  let url = `/repos/{owner}/{repo}/pulls/{pull_number}/commits?per_page=${limit}&page=${page}`;
  let result = await github.request(url, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: prNumber,
  });

  while (result.data.length !== 0) {
    dataSize = result.data.length;
    for (let i = 0; i < dataSize; i++) {
      listCommitInfo.push({
        sha: result.data[i].sha,
        message: result.data[i].commit.message,
      });
    }

    if (dataSize < limit) {
      break;
    }

    page++;

    url = `/repos/{owner}/{repo}/pulls/{pull_number}/commits?per_page=${limit}&page=${page}`;
    result = await github.request(url, {
      owner: context.repo.owner,
      repo: context.repo.repo,
      pull_number: prNumber,
    });
  }
  return listCommitInfo;
}
