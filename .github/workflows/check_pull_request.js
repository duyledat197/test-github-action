const regex = /\bLT-\d{1,6}\b/;

function getTicketList(data) {
  return data
    .map((el) => {
      if (el && el.commit && el.commit.message) {
        const val = el.commit.message.match(regex);
        if (val && val.length > 0) return val[0];
      }
      return null;
    })
    .filter((el) => el);
}

async function getPRCommits({ github, context, page = 1 }) {
  const result = await github.pulls.listCommits({
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: context.payload.number,
    page,
    per_page: 1,
  });

  if (result && result.data && result.data.length > 0) {
    return [...getTicketList(result.data), ...getPRCommits({ github, context, page: page + 1 })];
  }
  return getTicketList(result.data);
}

module.exports = {
  getPRCommits,
};
