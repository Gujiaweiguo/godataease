---
name: openspec-archive-change
description: Archive a completed change in the experimental workflow. Use when the user wants to finalize and archive a change after implementation is complete.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.1.1"
---

Archive a completed change in the experimental workflow.

**Input**: Optionally specify a change name. If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for available changes.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `openspec list --json` to get available changes. Use the **AskUserQuestion tool** to let the user select.

   Show only active changes (not already archived).
   Include the schema used for each change if available.

   **IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

2. **Check artifact completion status**

   Run `openspec status --change "<name>" --json` to check artifact completion.

   Parse the JSON to understand:
   - `schemaName`: The workflow being used
   - `artifacts`: List of artifacts with their status (`done` or other)

   **If any artifacts are not `done`:**
   - Display warning listing incomplete artifacts
   - Use **AskUserQuestion tool** to confirm user wants to proceed
   - Proceed if user confirms

3. **Check task completion status**

   Read the tasks file (typically `tasks.md`) to check for incomplete tasks.

   Count tasks marked with `- [ ]` (incomplete) vs `- [x]` (complete).

   **If incomplete tasks found:**
   - Display warning showing count of incomplete tasks
   - Use **AskUserQuestion tool** to confirm user wants to proceed
   - Proceed if user confirms

    **If no tasks file exists:** Proceed without task-related warning.

4. **Run test gate (unit + integration + E2E)**

    Before archiving, ensure all tests pass. This is a HARD REQUIREMENT.

    **4a. Check CI status for current branch**

    Run `gh pr list --head <current-branch> --json number,statusCheckRollup` to find the PR for this change.
    If no PR exists, warn the user and ask to create one first.

    Check if all CI checks passed:
    - If all passed: Note "CI checks passed"
    - If any failed: Block archive, tell user to fix CI failures first
    - If pending: Offer to wait or cancel

    **4b. Trigger and verify E2E tests**

    E2E tests are required before archive but not run on every PR.

    ```bash
    gh workflow run test-gate.yml \
      -f run_unit=false \
      -f run_integration=false \
      -f run_e2e=true \
      --ref <current-branch>
    ```

    Wait for the workflow to complete:
    ```bash
    gh run watch <run-id> --exit-status
    ```

    **If E2E tests fail:**
    - Block archive with clear error message
    - Show failure details from the workflow run
    - Tell user to fix issues and retry

    **If user wants to skip E2E:**
    - Use **AskUserQuestion tool** to confirm
    - Only allow skip with explicit acknowledgment of risk
    - Add warning to archive summary

5. **Assess delta spec sync state**

    Check for delta specs at `openspec/changes/<name>/specs/`. If none exist, proceed without sync prompt.

    **If delta specs exist:**
    - Compare each delta spec with its corresponding main spec at `openspec/specs/<capability>/spec.md`
    - Determine what changes would be applied (adds, modifications, removals, renames)
    - Show a combined summary before prompting

    **Prompt options:**
    - If changes needed: "Sync now (recommended)", "Archive without syncing"
    - If already synced: "Archive now", "Sync anyway", "Cancel"

    If user chooses sync, execute /opsx-sync logic (use the openspec-sync-specs skill). Proceed to archive regardless of choice.

6. **Perform the archive**

   Create the archive directory if it doesn't exist:
   ```bash
   mkdir -p openspec/changes/archive
   ```

   Generate target name using current date: `YYYY-MM-DD-<change-name>`

   **Check if target already exists:**
   - If yes: Fail with error, suggest renaming existing archive or using different date
   - If no: Move the change directory to archive

   ```bash
   mv openspec/changes/<name> openspec/changes/archive/YYYY-MM-DD-<name>
   ```

7. **Display summary**

   Show archive completion summary including:
   - Change name
   - Schema that was used
   - Archive location
   - Whether specs were synced (if applicable)
   - Note about any warnings (incomplete artifacts/tasks)

**Output On Success**

```
## Archive Complete

**Change:** <change-name>
**Schema:** <schema-name>
**Archived to:** openspec/changes/archive/YYYY-MM-DD-<name>/
**Specs:** ✓ Synced to main specs (or "No delta specs" or "Sync skipped")
**Tests:** ✓ CI passed, E2E passed (or "E2E skipped with acknowledgment")

All artifacts complete. All tasks complete.
```

**Guardrails**
- Always prompt for change selection if not provided
- Use artifact graph (openspec status --json) for completion checking
- Don't block archive on warnings - just inform and confirm
- Preserve .openspec.yaml when moving to archive (it moves with the directory)
- Show clear summary of what happened
- If sync is requested, use openspec-sync-specs approach (agent-driven)
- If delta specs exist, always run the sync assessment and show the combined summary before prompting
- **Test Gate (HARD REQUIREMENT)**: CI must pass and E2E tests must pass before archive
- Only allow E2E skip with explicit user acknowledgment
