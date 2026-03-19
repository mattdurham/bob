# First Principles Framework for Architecture Introspection

<!--
Attribution: This framework is adapted from the architecture-introspector skill by mahidalhan.
Original source: https://github.com/mahidalhan/claude-hacks/tree/master/skills/architecture-introspector
License: See original repository for license terms.
-->

## SpaceX Engineering Principles

### 1. Question Every Requirement
**Principle**: Every requirement should come with a name attached. Never accept "it's always been done that way" or "the regulation requires it" without verifying.

**Application to Architecture**:
- Trace every architectural decision to its original author/reason
- Question inherited patterns that may be cargo-culted
- Validate assumptions against current business needs
- Distinguish between hard constraints (laws of physics/business) and soft constraints (preferences, legacy patterns)

**Questions to Ask**:
- Who introduced this pattern and why?
- What problem was it originally solving?
- Does that problem still exist?
- What would happen if we removed this?

### 2. Delete the Part or Process
**Principle**: If you're not occasionally adding things back in, you're not deleting enough. The most common error is optimizing something that shouldn't exist.

**Application to Architecture**:
- Identify and remove architectural elements before optimizing them
- Look for layers of abstraction that don't carry their weight
- Remove components with single consumers that could be inlined
- Eliminate duplicated logic and redundant patterns

**Questions to Ask**:
- What happens if we remove this layer entirely?
- Is this abstraction solving a problem we actually have?
- How many consumers need this component? (If <2, consider inlining)
- What percentage of this module's features are actually used?

### 3. Simplify and Optimize
**Principle**: Only after removing unnecessary parts should you simplify and optimize what remains. Don't optimize something that shouldn't exist.

**Application to Architecture**:
- Simplify only after deletion phase is complete
- Reduce cognitive complexity in remaining components
- Consolidate similar patterns into unified approaches
- Optimize for clarity first, performance second

**Questions to Ask**:
- Can this be expressed more simply?
- Are we using the simplest tool for the job?
- Can we reduce the number of concepts developers need to learn?
- What's the cognitive load of understanding this?

### 4. Accelerate Cycle Time
**Principle**: Speed up the process, but only after the first three steps. Going faster makes bad processes worse.

**Application to Architecture**:
- Reduce time from change to production
- Eliminate unnecessary approval gates
- Automate what remains after deletion and simplification
- Reduce coupling to enable parallel development

**Questions to Ask**:
- How long from code change to production?
- What slows down feature development?
- Which architectural decisions create bottlenecks?
- Can teams deploy independently?

### 5. Automate
**Principle**: Automate last. Automating unnecessary complexity is worse than manual unnecessary complexity.

**Application to Architecture**:
- Automate only validated, simplified processes
- Build tooling for patterns that survived deletion
- Create guardrails for architectural principles
- Use code generation for boilerplate that shouldn't exist but must

**Questions to Ask**:
- Have we deleted and simplified first?
- Is this process mature enough to automate?
- Will automation hide problems we should fix instead?
- Does automation reduce or increase system complexity?

## Software Modularity Principle

### Core Tenet
**"Modularity without reuse is bureaucracy"**

Premature abstraction and excessive decomposition create overhead without benefit. Extract and modularize based on actual needs, not hypothetical futures.

### Modularity Guidelines

#### When to Extract (✅)
1. **Multiple Consumers** - 2+ components need the same logic
2. **Clear Boundaries** - The module has a well-defined, single responsibility
3. **Stable Interface** - The contract between modules is unlikely to thrash
4. **Independent Evolution** - The module can change without affecting internals of consumers

#### When NOT to Extract (❌)
1. **Single Consumer** - Only one component uses the logic (keep it cohesive)
2. **Tight Coupling** - The "module" is deeply intertwined with its consumer
3. **Premature Generalization** - Trying to make something reusable before understanding its patterns
4. **Unclear Boundaries** - The module's responsibility is fuzzy or overlaps with others

### Cohesion vs. Decomposition

**High Cohesion** (Good):
- Related functionality stays together
- Changes to a feature affect localized code
- Easy to understand and modify as a unit

**Excessive Decomposition** (Bad):
- Logic scattered across many small files
- Simple changes require touching multiple modules
- Cognitive overhead of navigating abstractions

### The 2-3 Rule
- **Services/Helpers**: Extract when 2+ consumers exist
- **Hooks/Utilities**: Extract when 3+ uses appear
- **Before these thresholds**: Keep code cohesive and localized

## Architecture Introspection Process

### Phase 1: Map Current State
1. Identify all architectural layers and components
2. Count consumers for each abstraction
3. Trace data/control flow paths
4. Document original decisions and their authors

### Phase 2: Apply SpaceX Principle 1 (Question)
1. For each component, ask: "Who required this and why?"
2. Validate assumptions against current reality
3. Identify cargo-culted patterns
4. Separate hard constraints from preferences

### Phase 3: Apply SpaceX Principle 2 (Delete)
1. Identify candidates for deletion:
   - Abstractions with <2 consumers
   - Layers adding no value
   - Duplicated logic
   - Unused features
2. Delete boldly (can always add back if needed)
3. Inline single-use abstractions
4. Remove unnecessary indirection

### Phase 4: Apply SpaceX Principle 3 (Simplify)
1. For components that survived deletion:
   - Reduce complexity
   - Unify similar patterns
   - Improve naming and contracts
   - Lower cognitive load

### Phase 5: Apply SpaceX Principle 4 (Accelerate)
1. Reduce coupling to enable parallel work
2. Eliminate architectural bottlenecks
3. Shorten feedback loops
4. Enable incremental deployment

### Phase 6: Apply SpaceX Principle 5 (Automate)
1. Identify validated, simplified patterns
2. Create tooling/guards for principles
3. Automate quality checks
4. Generate boilerplate for necessary patterns

## Output Format

### Architecture Analysis Report
```markdown
# Architecture Introspection: [System/Feature Name]

## Executive Summary
- Current state overview
- Key findings
- Recommended changes

## Phase 1: Current State Map
### Components Inventory
| Component | Consumers | Purpose | Original Author/Reason |
|-----------|-----------|---------|------------------------|

### Data Flow Diagram
[Visual or textual representation]

## Phase 2: Requirement Validation (Question)
### Challenged Assumptions
- [Assumption] → [Validation Result] → [Decision]

### Cargo-Culted Patterns
- [Pattern] → [Original Reason] → [Current Relevance]

## Phase 3: Deletion Candidates (Delete)
### Recommended Deletions
| Component | Reason | Impact | Risk |
|-----------|--------|--------|------|

### Inline Opportunities
- [Abstraction] → [Single Consumer] → [Inline Strategy]

## Phase 4: Simplification Plan (Simplify)
### Complexity Reduction
- [Component] → [Current Complexity] → [Simplified Approach]

### Pattern Unification
- [Multiple Patterns] → [Unified Pattern]

## Phase 5: Acceleration Opportunities (Accelerate)
### Bottleneck Removal
- [Bottleneck] → [Impact] → [Solution]

### Coupling Reduction
- [Tight Coupling] → [Decoupling Strategy]

## Phase 6: Automation Candidates (Automate)
### Validated Patterns for Automation
- [Pattern] → [Automation Approach] → [Expected Benefit]

## Implementation Plan
1. [Phase] - [Actions]
2. [Phase] - [Actions]

## Metrics
- Lines of Code: Before → After
- Component Count: Before → After
- Average Coupling: Before → After
- Deployment Time: Before → After
```

## Common Anti-Patterns to Identify

### 1. Premature Abstraction
- Framework for single use case
- Generic utilities never reused
- Abstraction layers with no consumers

### 2. Enterprise Fizz-Buzz
- 15 files to add two numbers
- Factories creating factories
- Managers managing managers

### 3. Cargo Cult Architecture
- "We use microservices because Google does"
- "Everything must be pluggable"
- "We need a service mesh for 3 services"

### 4. Resume-Driven Development
- Latest framework for simple problems
- Over-engineering to learn new tech
- Complexity for complexity's sake

### 5. Not-Invented-Here Syndrome
- Rebuilding existing solutions poorly
- Custom frameworks over proven libraries
- Reinventing wheels instead of using them

## Decision Framework

For every architectural element, answer:

1. **Who** introduced this and **why**? (Question)
2. **Can** we delete it? **Should** we? (Delete)
3. **If it stays**, how do we simplify it? (Simplify)
4. **Does it** slow us down? How to accelerate? (Accelerate)
5. **Should** we automate it? (Only after 1-4)

## Success Criteria

A successful architecture introspection achieves:

- ✅ Fewer components/layers (deletion complete)
- ✅ Clear, justified reasons for what remains (requirements validated)
- ✅ High cohesion within modules (proper boundaries)
- ✅ Low coupling between modules (independent evolution)
- ✅ Fast feedback loops (accelerated cycles)
- ✅ Minimal manual toil (automated where appropriate)
- ✅ Easy onboarding (reduced cognitive load)
- ✅ Every abstraction earns its keep (2+ consumers or compelling reason)
