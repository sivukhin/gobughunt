SELECT repo_id,
       repo_git_url,
       repo_git_branch,
       repo_last_git_commit_hash
FROM repos
WHERE repo_id = $1;