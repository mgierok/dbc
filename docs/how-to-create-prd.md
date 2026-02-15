# How to Create a High-Quality PRD from a Short Prompt

## 1. Purpose
Use this guide when an AI agent must create a Product Requirements Document (PRD) from a short user prompt.

The PRD must be:
- written in English,
- focused on product and business outcomes,
- a single source of truth for the feature scope,
- explicit about the change from current state to target state after release.

## 2. Core Rules (Non-Negotiable)
1. Use English only for all agent outputs in this workflow: clarifying questions, assumptions, answer summaries, feedback requests, quality checks, and PRD content.
2. Ask clarifying questions before drafting the PRD.
3. Focus on user value, business impact, measurable outcomes, and scope boundaries.
4. Do not include technical implementation details.
5. The goal is to produce one final PRD file; clarification and validation artifacts are allowed only to define and validate that PRD. Do not start implementation planning or execution.
6. If non-critical information is missing, capture it in `Assumptions` with confidence level.
7. Do not invent unknown constraints.
8. Do not finalize the PRD with any `TBD` values.
9. Do not include an `Open Questions` section or unresolved decision placeholders.
10. Continue clarification until all critical unknowns are resolved into explicit decisions or explicit scope exclusions.
11. Do not include metadata blocks, user stories, timeline, milestones, or change log.

## 3. Required Workflow (Execution Order)
Follow this sequence exactly:

1. Understand the request.
   - Restate the feature in one sentence.
   - Identify missing information.
2. Run mandatory clarification.
   - Start with 3-5 highest-impact questions.
   - Ask one question at a time (single-question mode).
   - Wait for the answer before asking the next question.
   - For each question provide exactly four options:
     - `A`, `B`, `C`: plausible and distinct options aligned with feature context and project context.
     - `D`: `Other: <User-defined answer>`.
   - If the answer is unclear, ask one follow-up clarification before moving on.
   - If two consecutive answers are still ambiguous, pause and ask one explicit decision question with trade-offs.
   - After the initial 3-5 questions, run additional targeted questions if needed to resolve remaining critical unknowns.
   - Do not move to final PRD output while any critical unknown remains unresolved.
3. Handle incomplete information.
   - Continue with an `Assumptions` section for non-critical items only.
   - Add confidence per assumption: `High`, `Medium`, or `Low`.
   - If a critical item cannot be resolved, do not produce final PRD. Ask a direct decision question and wait for answer.
4. Publish `Answer Summary`.
   - List confirmed answers and assumptions.
5. Draft the PRD using the fixed structure (Section 5).
6. Run one review loop.
   - Ask for focused feedback on scope, metrics, non-goals, and residual risks.
   - Revise if feedback is provided.
7. Run `Quality Gate Check`.
8. If any gate fails, revise and re-check before final output.

## 4. Clarifying Questions

### 4.1 Question Pool (Start with 3-5, then continue if needed)
Prioritize uncertainties that have the biggest impact on scope, business value, or success measurement, then ask follow-up questions until all critical unknowns are resolved.

1. Problem and context
   - What exact problem is being solved?
   - Why is it important now?
2. Target audience
   - Who is the primary user segment?
   - Who is explicitly out of scope?
3. Business outcomes
   - Which KPI should improve?
   - What impact magnitude is expected?
4. Current state
   - How is the problem handled today?
   - What pain points exist now?
5. Scope boundaries
   - What must be included in this release?
   - What is explicitly excluded?
6. Success criteria
   - Which metric, baseline, target, and time window define success?
   - What is the minimum acceptable release criterion?
7. Risks and dependencies
   - What risks could block value delivery?
   - What dependencies exist (teams, process, legal, data)?
8. Rollout and adoption
   - What behavior change is expected?
   - What may block adoption after release?

### 4.2 Mandatory Question Format
Use this exact structure for each question:

1. `Question`: one clear and specific question.
2. `Suggested answers`:
   - `A)` likely answer aligned with feature context and project context.
   - `B)` likely answer aligned with feature context and project context.
   - `C)` likely answer aligned with feature context and project context.
   - `D)` `Other: <User-defined answer>`.

Rules:
- `A`, `B`, and `C` must be plausible, mutually distinct, and context-aligned.
- `D` must always exist.
- Send only one question per message.
- Do not send question `N+1` before processing the answer to question `N`.

## 5. Fixed PRD Structure (Must Be Exact)
Use these headings in this exact order:

1. `Overview`
2. `Problem Statement`
3. `Current State (As-Is)`
4. `Target State After Release (To-Be)`
5. `Business Rationale and Strategic Fit`
6. `Goals`
7. `Non-Goals`
8. `Scope (In Scope / Out of Scope)`
9. `Functional Requirements`
10. `Non-Functional Product Requirements`
11. `Success Metrics and Release Criteria`
12. `Risks and Dependencies`
13. `Assumptions`

## 6. Section Rules (Authoring Guide)
1. `Overview`
   - Provide short feature summary and business value.
   - Include this problem-first hypothesis:
     - `We believe that [change] for [target segment] will [business/user outcome].`
     - `We will know this is true when [metric target] within [time window].`
2. `Problem Statement`
   - Describe current pain, affected users, and business consequences.
3. `Current State (As-Is)`
   - Explain how the product/process works now and where it fails.
4. `Target State After Release (To-Be)`
   - Describe expected future behavior and observable outcomes.
5. `Business Rationale and Strategic Fit`
   - Explain why this matters now and how it supports strategy.
6. `Goals`
   - Include only outcome-oriented goals.
7. `Non-Goals`
   - Define explicit boundaries to prevent scope creep.
8. `Scope (In Scope / Out of Scope)`
   - Describe release boundaries in plain business language.
9. `Functional Requirements`
   - Describe what the product must do from user/business perspective.
   - Use stable IDs: `FR-001`, `FR-002`, ...
   - Keep requirements atomic (one behavior per requirement).
   - Add one observable acceptance statement per `FR-*`.
10. `Non-Functional Product Requirements`
   - Define product quality needs in business terms (for example usability, reliability, compliance).
   - Use stable IDs: `NFR-001`, `NFR-002`, ...
11. `Success Metrics and Release Criteria`
   - Use quantified metrics and minimum pass criteria.
   - For each metric include: baseline, target, measurement window, and measurement method.
   - Include:
     - one `Primary Outcome Metric`,
     - 1-3 `Leading Indicators`,
     - at least one `Guardrail Metric`.
12. `Risks and Dependencies`
   - List risks to value delivery and key dependencies that can affect scope, release readiness, or metric outcomes.
13. `Assumptions`
   - List explicit assumptions used for missing input.

## 7. Quality Gates (All Must Pass Before Final Output)
1. Completeness
   - All 13 required sections exist and are non-empty.
2. Delta clarity
   - `Current State (As-Is)` and `Target State After Release (To-Be)` are both present and materially different.
3. Measurability
   - Every goal has at least one metric with baseline, target, and measurement window.
4. Scope control
   - Both in-scope and out-of-scope are explicit.
5. Requirement quality
   - Requirements are clear, atomic, outcome-focused, uniquely identified, and non-overlapping.
   - Every `FR-*` has one observable acceptance statement.
6. Business focus
   - No architecture, stack, API, schema, infrastructure, or implementation-plan content.
7. Assumption transparency
   - Missing information is visible in `Assumptions`.
8. No open questions policy
   - Final PRD contains no `Open Questions` section and no unresolved decision placeholders.
9. Clarification option quality
   - Each clarifying question has exactly `A`, `B`, `C`, `D`.
   - `A-C` fit feature and project context.
   - `D` allows user-defined input.
10. Clarification sequencing
   - Questions were asked one by one and answered before proceeding.
11. Clarification prioritization
   - Initial clarification round has 3-5 questions focused on highest-impact uncertainty.
   - Additional targeted questions are allowed and required when needed to resolve critical unknowns.
12. Metric quality
   - Every success metric includes baseline, target, measurement window, and method.
   - Metrics include one primary outcome metric, leading indicators, and guardrail metric.
13. Constraint integrity
   - Final PRD contains no `TBD`.
   - Every previously unknown critical constraint is resolved as a confirmed decision or explicit out-of-scope item.
14. Consistency
   - Goals, scope, requirements, and metrics do not contradict each other.
15. Review loop
   - Focused feedback was requested on scope, metrics, non-goals, and residual risks.
16. File output compliance
   - Final PRD is saved in `.tasks` as `PRD[id]-[short-name].md` with next numeric `id`.

## 8. Anti-Patterns (Reject and Revise)
Reject the draft if any of these appear:

1. Generic statements not tied to user or business outcomes.
2. Vague success criteria without measurable targets.
3. Duplicate or overlapping requirements.
4. Weak or missing `Non-Goals`.
5. Hidden assumptions (not listed in `Assumptions`).
6. Technical solution details mixed with product requirements.
7. Solution-first framing without clear problem evidence.
8. Vanity metrics without decision value.
9. Scope definition without explicit trade-offs.
10. Any violation of Section 2 rules `8-10` (for example `TBD`, `Open Questions`, unresolved decision placeholders).

## 9. Forbidden Content
Do not include:
- technical architecture or code-level design,
- database schema or API contract details,
- sprint or task breakdown,
- implementation steps,
- metadata block,
- user stories section,
- timeline or milestones section,
- change log/history section,
- unresolved placeholders (see Section 2 rules `8-10`).

## 10. File Output Rules
When saving the generated PRD:

1. Save in `.tasks`.
2. Use filename format `PRD[id]-[short-name].md`.
3. Set `[id]` to next numeric ID among `.tasks/PRD*-*.md`.
4. Use a short, meaningful `[short-name]`.
5. Use kebab-case for `[short-name]`.

## 11. Agent Output Contract
When generating a PRD, follow this output flow:

1. Start clarification in single-question mode with `Question 1` only.
2. Wait for user response, then continue to `Question 2`, and so on.
3. Continue asking targeted questions until all critical unknowns are resolved.
4. After clarification, output `Answer Summary` with confirmed choices and assumptions.
5. Output the PRD using the exact fixed structure.
6. Request focused feedback on scope, metrics, non-goals, and residual risks.
7. Revise draft if feedback is provided.
8. Save final PRD to `.tasks` using required naming (Section 10).
9. End with `Quality Gate Check` and mark each gate `PASS`.
10. If any gate is not `PASS`, revise and repeat the check before finalizing.
