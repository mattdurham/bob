---
name: bob-matt-grilling
description: Grill the user relentlessly about a plan, decision, or idea. Use when the user wants to stress-test their thinking, or uses any 'grill' trigger phrases.
metadata:
  author: Matt Pocock
  source: https://github.com/mattpocock/skills/tree/49dd158d1076134a641b33efb035946536778336/skills/productivity/grilling
  upstream-commit: "49dd158d1076134a641b33efb035946536778336"
license: MIT
---

Interview the user until you reach shared understanding of the plan, decision,
or idea. Map the decisions as a design tree and use the existing context so
settled decisions are not repeated. The frontier contains decisions whose
prerequisites are already settled. Choose the most useful available decision and
ask exactly one focused question at a time; do not bundle independent questions
or present an entire round. State your recommended answer so a simple yes can
accept it, then wait for the user’s answer before asking the next question.

Each answer reshapes the tree: record the settled decision, recompute the
frontier, and ask the next question whose prerequisites are known. Questions
that depend on an unanswered decision must wait. Finding facts is your job:
inspect the environment or dispatch a bounded fact-finding subagent when useful
rather than asking the user for information you can look up. A pending
exploration leaves its dependent decisions unsettled; you may ask one other
ready question while that fact is being investigated. Decisions remain the
user’s, and every question requires an answer before the interview continues.

Finish when every material branch is settled and no assumption has been silently
treated as a decision. Summarize the agreed understanding, identify any
remaining uncertainty, and ask for confirmation before acting on it. Preserve
previously confirmed decisions when the interview resumes, revisiting them only
when new evidence or the user’s correction changes their premises.
