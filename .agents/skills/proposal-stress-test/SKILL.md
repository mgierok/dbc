---
name: proposal-stress-test
description: Critically evaluate a proposed repository change or an existing implementation plan by stress-testing assumptions, consequences, risks, edge cases, and execution gaps across product, technical, operational, and delivery concerns. Use only when the user explicitly requests `proposal-stress-test` by name to challenge, de-risk, or refine a product or technical idea before implementation, or to review a concrete plan with a critical eye before execution.
---

# Proposal Stress Test

## Purpose

This skill performs analysis-only review before implementation starts.

The skill MUST help the user:
- make the proposal or plan explicit and testable,
- surface hidden assumptions,
- identify risks, constraints, tradeoffs, and edge cases,
- understand likely downstream consequences,
- decide whether the idea is ready for implementation, needs refinement, or should be rejected.

## Explicit Invocation Policy

- This skill MUST run only when the user explicitly requests `proposal-stress-test` by name.
- This skill MUST NOT auto-trigger.
- If the user asks for this style of review without naming the skill, the agent MUST ask whether they want to invoke `proposal-stress-test`.

## Scope

- In scope:
  - product ideas,
  - technical changes,
  - architecture changes,
  - workflow or process changes inside the repository,
  - concrete implementation plans that are intended to be executed,
  - mixed proposals that affect both product and engineering behavior.
- Out of scope:
  - implementing the proposal,
  - writing production code,
  - writing final specifications unless the user explicitly asks for them after the review.

## Core Behavior

- The skill MUST start by understanding what the user actually means before criticizing the idea.
- The skill MUST detect which input mode applies:
  - `idea mode`: the user describes what they want to do,
  - `plan mode`: the user provides or references a concrete plan to review.
- In `idea mode`, the skill MUST ask concrete clarification questions when the proposal is ambiguous, incomplete, or mixes multiple concerns.
- In `plan mode`, the skill MUST review the plan itself as the main artifact under critique instead of asking broad discovery questions about the original idea.
- The skill MUST challenge assumptions instead of accepting the proposal framing at face value.
- The skill MUST prefer evidence-based concerns over generic pessimism.
- The skill MUST distinguish clearly between:
  - confirmed consequences,
  - likely risks,
  - open questions,
  - assumptions made due to missing information.
- The skill MUST optimize for decision quality, not for politeness-only agreement.
- The skill SHOULD be direct, specific, and technically grounded.
- The skill MUST NOT derail the conversation with speculative objections that have no plausible impact.

## Input Modes

### Idea Mode

- `Idea mode` applies when the user provides a prompt, rough concept, or desired change without a concrete execution plan.
- In `idea mode`, the skill MUST clarify the proposal first, then stress-test it.

### Plan Mode

- `Plan mode` applies when the user provides a concrete plan or points to a plan artifact that is intended for execution.
- In `plan mode`, the skill MUST treat the plan as already-defined scope.
- In `plan mode`, the skill MUST NOT ask broad questions whose only purpose is to refine the original idea at a higher level.
- In `plan mode`, the skill SHOULD ask questions only when the plan is missing decision-critical execution detail or contains ambiguity that blocks meaningful review.
- In `plan mode`, the skill MUST evaluate whether the plan is:
  - complete enough to execute safely,
  - internally consistent,
  - aligned with repository constraints,
  - missing risks, sequencing, validation, migration, or rollback thinking.

## Clarification Workflow

1. In `idea mode`, the skill MUST restate the proposal in concrete terms.
2. In `idea mode`, the skill MUST identify missing information that blocks sound evaluation.
3. In `idea mode`, the skill MUST ask focused clarification questions before giving a final assessment when key ambiguity remains.
4. In `idea mode`, clarification questions MUST target decision-critical gaps such as:
   - user impact,
   - problem being solved,
   - success criteria,
   - affected workflows,
   - boundaries of change,
   - compatibility expectations,
   - rollout or migration constraints,
   - non-goals.
5. In `idea mode`, the skill SHOULD ask the minimum number of questions needed to remove major ambiguity.
6. In `idea mode`, if the user provides partial information only, the skill MAY continue with explicit assumptions, but those assumptions MUST be labeled.
7. In `plan mode`, the skill MUST restate the plan briefly before critique.
8. In `plan mode`, the skill MUST focus on gaps in the plan rather than on reconstructing the original problem statement from scratch.

## Stress-Test Areas

The skill MUST evaluate only the areas relevant to the proposal.

1. Problem framing
- Is the proposal solving the real problem or only a symptom?
- Is the current pain concrete, recurring, and worth the change cost?

2. User and workflow impact
- Who benefits, who pays the cost, and who needs to change behavior?
- Which existing workflows become slower, harder, or more confusing?

3. Product semantics and correctness
- Could the change create ambiguous behavior, hidden rules, or inconsistent UX?
- Are there edge cases where the proposed behavior becomes surprising or contradictory?

4. Technical design and architecture
- Does the proposal fit current boundaries and dependency direction?
- Could it create coupling, leakage, or future maintenance hotspots?

5. Data, state, and migration risk
- Does the change affect persistence, backward compatibility, existing records, or state transitions?
- Are there partial-migration or rollback hazards?

6. Operational and delivery risk
- Does the proposal increase complexity in rollout, support, observability, testing, or incident handling?
- Are there hidden runtime costs, failure modes, or performance risks?

7. Security and trust boundaries
- Could the change weaken validation, permissions, isolation, or auditability?

8. Long-term maintainability
- Does this create a one-off rule, special case, or policy exception that will keep spreading?
- Is there a simpler alternative with lower blast radius?

9. Plan quality and execution safety
- Are responsibilities, order of work, and dependencies clear?
- Does the plan omit validation, testing, observability, migration, rollback, or release steps?
- Are any tasks underspecified, unjustified, or likely to produce rework?

## Analysis Workflow

1. Detect whether `idea mode` or `plan mode` applies.
2. In `idea mode`, clarify the proposal first.
3. In `idea mode`, identify the intended upside and the problem being solved.
4. In both modes, list major assumptions behind the proposal or plan.
5. Stress-test the proposal or plan across relevant `Stress-Test Areas`.
6. Identify:
   - likely benefits,
   - likely costs,
   - risks and failure modes,
   - edge cases,
   - hidden implementation or migration complexity,
   - alternative approaches worth considering.
7. In `plan mode`, additionally identify:
   - missing tasks,
   - bad sequencing,
   - weak acceptance criteria,
   - hidden dependencies,
   - verification gaps,
   - rollback or release gaps.
8. Conclude with a decision-oriented assessment.

## Output Contract

When `idea mode` input is still too vague, the response MUST contain:
- a short restatement of current understanding,
- a short list of concrete clarification questions,
- no fake certainty.

When `idea mode` input is clear enough to assess, the response MUST contain these sections in this order:

1. `Proposal`
- concise restatement of the idea or plan being reviewed.

2. `What Looks Reasonable`
- the strongest valid arguments in favor of the proposal.

3. `Concerns`
- the most important problems, tradeoffs, and unintended consequences.

4. `Edge Cases`
- specific scenarios where the proposal may fail, confuse users, or complicate the system.

5. `Open Questions`
- unresolved issues that should be answered before implementation.

6. `Recommendation`
- one of:
  - `Proceed`,
  - `Proceed with changes`,
  - `Do not proceed yet`,
  - `Reject`.
- The recommendation MUST include a short reason.

When `plan mode` input is clear enough to assess, the response MUST contain these sections in this order:

1. `Proposal`
- concise restatement of the plan being reviewed.

2. `What Looks Reasonable`
- the strongest valid arguments in favor of the plan.

3. `Plan Gaps`
- missing steps, sequencing issues, unclear ownership, weak acceptance criteria, or verification holes.

4. `Concerns`
- the most important problems, tradeoffs, and unintended consequences.

5. `Edge Cases`
- specific scenarios where the plan may fail, cause rework, or miss important safeguards.

6. `Open Questions`
- unresolved issues that should be answered before execution.

7. `Recommendation`
- one of:
  - `Proceed`,
  - `Proceed with changes`,
  - `Do not proceed yet`,
  - `Reject`.
- The recommendation MUST include a short reason.

## Quality Bar

- The skill MUST surface the highest-impact concerns first.
- The skill MUST prefer concrete examples over abstract warnings.
- The skill SHOULD call out second-order effects, not only direct effects.
- The skill MUST note when a concern is an inference rather than a confirmed fact.
- The skill MUST NOT present implementation work as trivial unless that claim is defensible.
- The skill MUST NOT confuse disagreement with analysis; each major concern SHOULD explain why it matters.

## Example Triggers

- `Use proposal-stress-test: I want to merge onboarding and account settings into one flow.`
- `Run proposal-stress-test for this caching idea.`
- `Invoke proposal-stress-test. I want to move validation from domain to handler layer.`
- `Use proposal-stress-test on plan PRD-014/TASK-03.`
- `Run proposal-stress-test against this implementation plan before execution.`
