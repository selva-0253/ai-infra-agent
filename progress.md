# AI-Infra-Agent Progress Log

## Project Summary
**Goal:** Build an autonomous Infrastructure-as-Code agent that interprets natural language requests to provision and manage cloud resources.
**Current Phase:** GCP Migration & Clean Architecture Refactor

---

## Roadmap / Timeline

### Phase 1: MVP - AWS Lambda Implementation (Feb 2026 - Early April 2026)
* **Concept:** Designed a Go-based agent to run on AWS Lambda.
* **Key Integrations:** Integrated AWS Bedrock (Claude) to interpret natural language prompts into infrastructure actions.
* **Milestones:** * Successfully provisioned EC2 instances programmatically.
    * Resolved initial "Nil Pointer" runtime errors and AMI lookup issues.
    * Implemented basic VPC/Subnet logic.
* **Key Challenges:** Initial code was monolithic (`main.go`).

### Phase 2: Stabilization & Modularization (April 2026)
* **Refactoring:** Moved away from a "God File" to a modular structure (separating EC2, Bedrock logic).
* **Idempotency:** Implemented `ClientToken` logic in `RunInstances` to prevent redundant provisioning during retries.
* **Observability:** Fixed log flushing issues in Go/Lambda using `log.SetOutput(os.Stdout)`.
* **Identity:** Successfully migrated identity management logic to support AWS IAM roles.

### Phase 3: The Pivot - GCP & Clean Architecture (Late April 2026 - Present)
* **The Pivot:** Migrated from AWS Lambda to GCP Compute Engine.
    * *Reasoning:* Lambda's 15-minute timeout was a blocker for long-running reconciliation/agentic loops.
* **Architecture Shift:** Moving to "Clean Architecture" pattern (`cmd/`, `pkg/`, `internal/`).
* **New Integrations:** Swapping AWS Bedrock for Google Vertex AI (Gemini) SDK.
* **Infrastructure:** Commenced design of GCP-native provisioning logic (`Instances.Insert`).

---

## Architectural Decision Records (ADRs)

| ID | Title | Date | Status |
| :--- | :--- | :--- | :--- |
| ADR-001 | AWS Lambda to GCP Compute Engine | 2026-04-26 | Accepted |
| ADR-002 | Adoption of Clean Architecture (Provider Pattern) | 2026-04-27 | Accepted |

* *Note: See `/docs/adr/` for full details.*

---

## Next Steps (Kanban Board Priorities)
1.  **[Infrastructure]** Setup GCP Service Account with `Compute Admin` and `Vertex AI User` roles.
2.  **[Code]** Implement `CloudProvider` interface in `pkg/provider`.
3.  **[Code]** Complete migration of Bedrock client to Vertex AI (`genai` SDK).
4.  **[Ops]** Configure `systemd` service for persistent agent execution on GCE.

---

## Lessons Learned
* **Idempotency is non-negotiable:** Relying on `ClientToken` is the only safe way to handle distributed infrastructure requests.
* **Modularization reduces anxiety:** Separating the "Agent Brain" from the "Cloud Hands" prevents dependency-hell when switching cloud providers.
* **Build vs. Buy:** This project is worth it because it builds a **Domain-Specific Operator**, not just a coding assistant.