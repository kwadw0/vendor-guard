---
description: Update code graph after changes (graphify update + cluster)
---

Run graphify incremental update for the project and regenerate the report.

Execute these steps:

1. Run `graphify update .` to re-extract changed files and patch `graphify-out/graph.json` (use `GRAPHIFY_FORCE=1 graphify update . --force` if files were deleted/renamed).
2. Then run `graphify cluster-only .` to re-cluster and regenerate `graphify-out/GRAPH_REPORT.md` and `graphify-out/graph.html`.
3. Print the final `graphify-out/GRAPH_REPORT.md` summary (nodes, edges, communities) and confirm the update succeeded.

If graphify is not installed, inform the user to install with `pip install graphify` or `uv tool install graphify`.

$ARGUMENTS
