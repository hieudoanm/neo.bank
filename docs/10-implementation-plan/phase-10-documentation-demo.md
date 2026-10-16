# NEO Bank — Implementation Plan — Phase 10: Documentation and Demo

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Complete all documentation and record a demo of the complete workflow. Phase 10 ensures the system is fully documented and demonstrable before any cloud deployment.

**Week:** 7

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Complete documentation | All 12 normative docs reviewed and consistent |
| Demo script/video | Register → login → balance → transfer → status → history |
| README | Architecture diagram (Mermaid) and setup instructions |
| AGENTS.md | Contributor guidelines and project conventions reviewed and updated (pre-existing normative root doc) |
| Decision records | ADRs in `docs/decisions/` for significant architectural decisions |
| Consistency verification | All docs match the implementation |

## Tasks

- [ ] Review and update all documentation (00-12).
- [ ] Create architecture diagram (Mermaid) in README.
- [ ] Record demo showing: register → login → balance → transfer → status → history.
- [ ] Review and update AGENTS.md (pre-existing normative root doc) so project conventions match the implementation.
- [ ] Record significant decisions as ADRs in `docs/decisions/` per AGENTS.md delivery rules.
- [ ] Update README with clear setup instructions (clone, install, run).
- [ ] Verify all documentation is consistent with implementation (no drift).
- [ ] Verify all cross-references between docs resolve correctly.
- [ ] File numbering verified (00-12, consistent across all docs).

## Acceptance Criteria

- All documentation files complete, consistent, and under 300 lines each.
- Demo recorded and verified against the actual implementation.
- README includes architecture diagram and setup instructions.
- AGENTS.md defines project conventions followed by all code.
- Significant decisions are recorded as ADRs in `docs/decisions/`.
- Cross-references between docs resolve correctly (no broken links).
- All files numbered 00-12 consistently.

## Dependencies

Phase 9 (complete system with observability to document).

## Testing

- Documentation review: all normative docs present and current.
- Link verification: all internal doc links resolve.
- Demo verification: demo matches actual UI behavior.
- Consistency check: docs match code (API endpoints, schema, config).

## Risks & Considerations

- **Documentation drift:** docs must be updated whenever behavior changes.
- **Demo fidelity:** demo must use real API responses, not mocked/faked behavior.
- **File numbering:** consistent 00-12 numbering across all docs and references.

## Cross-references

- All normative docs: [`docs/00-overview.md`](../00-overview.md) through [`docs/12-operations.md`](../12-operations.md)
- AGENTS.md: `AGENTS.md`
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_10_documentation_demo` (verifies documentation completeness)