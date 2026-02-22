# Bob's Pirate Mode 🏴‍☠️

## Ahoy Matey!

Captain Bob now talks like a proper pirate throughout the workflows!

## What Changed

### 🎯 Main Workflows Updated

#### bob:work (Sequential Workflow)
- **Greeting**: "Ahoy matey! ⚓ Captain Bob at yer service!"
- **Phase transitions**: "Settin' sail to PLAN phase, matey..."
- **Status updates**: "REVIEW found 3 barnacles on the hull → routing to EXECUTE to scrape 'em off"
- **Completion**: "HOIST THE COLORS! Well done, matey!"

#### bob:team-work (Agent Teams Workflow)
- **Greeting**: "Ahoy matey! ⚓ Let me rally me crew of agent teammates!"
- **Team kickoff**: "All hands on deck, mateys! The voyage begins!"
- **Missing feature**: "Avast! 🏴‍☠️ Agent teams be locked in the hold, matey!"
- **Completion**: "Shiver me timbers! The code be battle-tested and ready to sail!"

#### brainstorming
- **Greeting**: "Ahoy matey! ⚓ Let's chart a course fer this idea o' yers!"

## Example Session

### Before (Boring):
```
Starting workflow...
Brainstorm complete
Proceeding to PLAN...
All checks passed. Ready to merge?
```

### After (Pirate!):
```
Ahoy matey! ⚓ Captain Bob at yer service!

Ye be wantin' to build: Add rate limiting

Let me chart a course through these waters!

---

⚓ BRAINSTORM complete → Chart marked at .bob/state/brainstorm.md
Settin' sail to PLAN phase, matey...

⚓ PLAN complete → Course plotted at .bob/state/plan.md
All hands to EXECUTE, matey...

---

Shiver me timbers! ⚓ All checks be passin', matey!

The code be battle-tested and ready to sail!

Shall we merge this fine work into the main ship? [yes/no]

---

🏴‍☠️ HOIST THE COLORS! 🏴‍☠️

Well done, matey! Ye've built yerself some fine code!
The treasure be safely stowed in the main ship!

May yer builds be swift and yer bugs be few!
Fair winds and following seas! ⚓

— Captain Bob
```

## Pirate Vocabulary Used

| Term | Meaning |
|------|---------|
| **Ahoy matey!** | Hello friend! |
| **Aye aye!** | Yes sir! / Understood! |
| **Avast!** | Stop! / Warning! |
| **Hoist the colors!** | Raise the flag / Celebrate! |
| **Shiver me timbers!** | Expression of surprise |
| **All hands on deck!** | Everyone get to work! |
| **Fair winds and following seas!** | Good luck and safe travels! |
| **Barnacles** | Problems/bugs in the code |
| **Treasure** | The work/tasks to complete |
| **Chart** | Plan/document |
| **Shipshape** | Clean and ready |
| **Seaworthy** | Good quality, ready to use |
| **The main ship** | Main branch |
| **Captain** | Team lead/orchestrator |
| **Crew** | Teammate agents |

## Team Workflow Messages

When using **bob:team-work**, the teammates also use pirate language:

**Coder completing task:**
```
"Aye captain! Task 123 be complete: Implement authentication
The code be shipshape and tests be passin'!"
```

**Coder encountering issue:**
```
"Avast, captain! Hit some rough waters on task 123: Nil pointer error
Need yer guidance to navigate through!"
```

**Reviewer approving:**
```
"Aye captain! Task 123 be approved!
The code be seaworthy and ready to sail! ⚓"
```

**Reviewer finding issues:**
```
"Avast, captain! Found some barnacles on task 123!
Created 3 fix tasks to scrape 'em off. Details in the task list."
```

**Reviewer finding critical issue:**
```
"ALL HANDS ON DECK! Critical issue in task 123!
SQL injection vulnerability in auth.go:42
This needs the captain's attention right away!"
```

## Status Updates

Throughout the workflow, you'll see pirate-themed status updates:

```
⚓ BRAINSTORM complete → Chart marked at .bob/state/brainstorm.md
Settin' sail to PLAN phase, matey...

⚓ PLAN complete → Course plotted at .bob/state/plan.md
All hands to EXECUTE, matey...

⚓ EXECUTE complete → Code be built!
Runnin' the TEST battery, matey...

⚓ TEST complete → All tests be passin'!
Sendin' the crew for REVIEW, matey...

⚓ REVIEW found 3 barnacles → Routing to EXECUTE to scrape 'em off

⚓ All clean! → Committin' to the ship's log, matey...
```

## Error Messages

Even errors are pirate-themed:

**Missing experimental flag:**
```
Avast! 🏴‍☠️ Agent teams be locked in the hold, matey!

Run this command to set 'em free:
  make enable-agent-teams

Then restart Claude Code and hoist the sails again!
```

## Installation

The pirate language is built into the skills, so just install normally:

```bash
make install
```

Then invoke any Bob workflow:
```bash
/bob:work "Add new feature"
/bob:team-work "Add new feature"
/brainstorming
```

## Turn Off Pirate Mode (If You Must)

If you need formal language for documentation or presentations, you can temporarily override by asking:

```
"Please use formal language for this session"
```

But why would ye want to, matey? 🏴‍☠️

## Files Modified

- `skills/work/SKILL.md` - Added pirate greetings and status updates
- `skills/team-work/SKILL.md` - Added pirate greetings, kickoff, and completion
- `skills/brainstorming/SKILL.md` - Added pirate greeting

## Summary

Captain Bob now talks like a proper pirate throughout all workflows!

Key additions:
- ⚓ Pirate greetings ("Ahoy matey!")
- ⚓ Pirate status updates ("Settin' sail to...")
- ⚓ Pirate completion messages ("HOIST THE COLORS!")
- ⚓ Pirate error messages ("Avast!")
- ⚓ Pirate teammate messages (team-work only)

**May yer builds be swift and yer bugs be few!**
**Fair winds and following seas! ⚓**

🏴‍☠️ **— Captain Bob, Belayin' Pin of Your Agents** 🏴‍☠️
