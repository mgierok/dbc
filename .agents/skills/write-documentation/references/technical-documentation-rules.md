# Technical Documentation Rules

## Target

- Target file: `docs/technical-documentation.md`
- Related template: `references/technical-documentation-template.md`

## Scope

Document implementation-facing facts only:

- architecture and boundaries,
- components and responsibilities,
- technical mechanisms and contracts,
- runtime/operational constraints and risks.

Do not define or restate business rules; keep focus on technical mechanisms and implementation behavior.

## Allowed Topics (Mandatory Allowlist)

Each proposed technical-documentation change MUST map to at least one topic below:

- System architecture, layers, and boundary definitions.
- Dependency direction rules and architectural invariants.
- Component/module/package responsibilities and ownership.
- Internal interfaces, contracts, and integration points.
- Data models, schema contracts, and data-flow definitions.
- Runtime lifecycle, initialization, and shutdown behavior.
- Execution flow orchestration and control flow mechanics.
- State management mechanisms and transition logic.
- Persistence/read/write mechanisms and transactional behavior.
- Error classification, handling strategy, and propagation paths.
- Resilience mechanisms (retries, fallbacks, idempotency, rollback).
- Concurrency, synchronization, and cancellation behavior.
- Security controls in implementation (authentication, authorization, validation, sanitization).
- Privacy/data-protection implementation constraints.
- Performance characteristics, bottlenecks, and optimization tradeoffs.
- Resource management (connections, memory, file handles, cleanup behavior).
- Observability implementation (logging, metrics, tracing, diagnostics).
- External dependency integration mechanics and adapter behavior.
- Configuration loading, precedence, and runtime configuration contracts.
- Environment/runtime operational constraints and failure modes.
- Backward-compatibility and migration mechanics.
- Testing strategy, technical coverage boundaries, and verification approach.
- Toolchain/runtime dependency versions relevant to implementation.
- Technical risks, limitations, and known tradeoffs.
- Deep-dive references for architecture/testing internals.

Changes that are primarily user-facing product narrative, UX copywriting, marketing language, or business-policy description without implementation relevance are out of scope for technical documentation.

## Out-of-Allowlist Handling (Mandatory)

If a proposed change does not map to any allowed technical topic:

- skip the change item, or
- adapt it to the nearest allowed technical topic while preserving factual meaning and without adding scope.

Never introduce technical-documentation content outside this allowlist.

## Topic Mapping Output (Mandatory)

When adding or changing technical-documentation information, explicitly indicate which allowed topic(s) the change maps to.

Minimum requirement:

- provide mapping per applied change item,
- use exact allowed-topic labels from this file,
- include this mapping in the completion output.

## Writing Rules

- Write for engineering and maintenance audiences.
- Describe current factual implementation state only.
- Keep content practical, code-aligned, and free from repeated product-level behavior narratives.
- Do not describe development workflow (branching, PR flow, delivery process).
- Preserve existing section numbering and anchors unless structural migration is explicitly requested.

## Consistency and Integrity Contract (Mandatory)

- Treat this rules file as the parent normative document for technical documentation writing.
- Treat `references/technical-documentation-template.md` as the structural contract for the technical document.
- Keep the generated/updated technical document structurally aligned with the template section set and order.
- If this rules file changes structural requirements, update the related template in the same change set.
- If the related template changes structure, update this rules file in the same change set.
- Do not leave contradictions between this rules file and the related template.
