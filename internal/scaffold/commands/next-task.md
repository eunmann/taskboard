---
description: Pick the next unblocked task, implement it, and mark it done
---

# Next Task

The core execution loop: pick one task, do it well, mark it done.

## 1. Find the Next Task

- Read all `.taskboard/*.yaml` files.
- Filter to tasks where `status: open` AND every `blocked-by` ref has `status: done` or `status: cancelled`.
- Sort by priority (`critical` > `high` > `medium` > `low`), then by `created` (earliest first) as tiebreaker.
- If no tasks are available, report "All tasks are blocked or complete" and stop.

## 2. Confirm

- Show the selected task's ID, title, priority, and description.
- Ask the user to confirm before proceeding.

## 3. Claim

- Set `status: in-progress` in the task YAML file.
- Update the `updated` timestamp to now.

## 4. Plan

- Read the task description and acceptance criteria.
- Identify which files and packages are affected.
- Outline the implementation steps.

## 5. Implement

- Write tests first when adding behavior.
- Implement one step at a time.
- After each step, verify: run tests and check for lint errors.
- Commit atomically after each verified step.

## 6. Complete

- Verify all acceptance criteria are met.
- Set `status: done` in the task YAML file.
- Update the `updated` timestamp to now.

## 7. Report

Summarize:
- What was implemented.
- Which tests were added or modified.
- Which tasks are now unblocked (their `blocked-by` refs are all resolved).

Suggest: "Run `/next-task` to continue, or `/task-status` for an overview."
