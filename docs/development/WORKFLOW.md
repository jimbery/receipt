# Development workflow

## Principle: test-first gates

Every milestone gate must be expressible as an **automated or scripted measurement**. If it can't be measured, refine the gate before writing production code.

## Cycle

### 1. Roadmap defines the phase

The [roadmap](../roadmap/ROADMAP.md) states the problem, gate, and kill/pivot criterion. No code yet.

### 2. Milestone doc triggers the build

Copy [`milestones/_template/MILESTONE.md`](../milestones/_template/MILESTONE.md) (or open the phase-specific doc) and complete:

- **Deliverables** — packages, files, interfaces
- **Test plan** — table tests, property/fuzz tests, fixture layout, harness metrics
- **Acceptance criteria** — numeric thresholds copied from the roadmap gate
- **ADRs required** — list open decisions; write ADRs before implementation

Set status to **Ready** only when the test plan is concrete enough that someone else could implement from it.

### 3. ADRs constrain the implementation

If the milestone lists required ADRs, those must be **accepted** before status moves to **In progress**.

### 4. Implement against tests

Order of work:

1. Extend the **canonical model** (`internal/model/`) if the phase adds fields
2. Write **failing tests** that encode acceptance criteria
3. Implement until tests pass
4. Run the **harness** (`internal/harness/`) if the phase affects coverage or match quality

### 5. Gate review

Run:

```bash
make test          # unit + race
make test-fuzz     # fuzz targets (short mode in CI)
make test-cover    # coverage report
```

Record results in the milestone doc **Gate review** section. Compare against acceptance criteria.

### 6. Close or pivot

- **Pass** → milestone status **Done**; update roadmap changelog if scope shifted
- **Fail gate, pass kill criterion** → record pivot in roadmap changelog; do not patch forward silently
- **Fail gate, mitigation exists** → e.g. accelerate OCR before judging itemisation; update milestone sequencing in roadmap

## Package conventions

```
internal/
  model/       # Canonical types — Transaction, Receipt, Match
  match/       # Matching engine (Phase 0+)
  harness/     # Coverage & quality metrics
  <phase>/     # Phase-specific packages as they land
test/
  testdata/    # Shared fixtures (JSON, synthetic labelled pairs)
```

- `internal/` packages are private to this module
- Anything that must be measured by the harness exposes pure functions over `model` types
- Integration tests live under `test/` or `*_integration_test.go` with build tags

## What blocks a merge

- Milestone acceptance tests not written or not passing
- Gate metric regression without ADR explaining the trade-off
- New external dependency without ADR
