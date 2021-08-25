async function getPRCommits() {
  const result = await github.pulls.listCommits({
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: context.payload.number,
  });
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

async function getTicketIDs(releaseDate, jiraUser, jiraToken) {
  const url = "https://manabie.atlassian.net/rest/api/3/search?jql=";
  const tmpl = escape(`project = LT AND summary ~ ${releaseDate} AND issuetype = Release`);

  const result = await fetch(`${url}${tmpl}`, {
    headers: {
      method: "GET",
      Authorization: `Basic ${base64.encode(jiraUser + ":" + jiraToken)}`,
      "Content-Type": "application/json",
      // 'Content-Type': 'application/x-www-form-urlencoded',
    },
  });
  console.log(result);
}

module.export = {
  getPRCommits,
  getTicketIDs,
};
