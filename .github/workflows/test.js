async function getPRCommits({ github, context }) {
  const result = await github.pulls.listCommits({
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: context.payload.number,
  });
  console.log(JSON.stringify(github.pulls));
  const regex = /\bLT-\d{1,6}\b/;
  return result.data
    .map((el) => {
      if (el && el.commit && el.commit.message) {
        const val = el.commit.message.match(regex);
        if (val && val.length > 0) return val[0];
      }
      return null;
    })
    .filter((el) => el);
}

module.exports = {
  getPRCommits,
};
