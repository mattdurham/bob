I want to create a folder in bob named bob, this will be a simple coding harness like pi.dev or crush. I want to use charmbracelet v2 for the core. I want extensions to be written in wasm compiled language using wazero with a /reload, and I want it to work with multuple providers.

in ~/source/pi-mono is pi.dev sourcecode for inspiration. Ideally I want a basic TUI, provider hooks and lifecycle hooks to be written in go code, and most other functionality exported via extensions. The SDK/API should have similiar scope to pi.dev. 

Goals
* Simple
* Extensible
* Core functionality that is very limited
* Core in go and charmbracelet
* Extensible via wasm/wazero
* Good Docs on the API
* Core functionality should be UI, handling the lifecycle and handling extensions. Ideally the most basic is you run the app and you get a system that can communicate with a provider but honestly not even be able to read and write files.

## Addendums

* Extensions should be able to load/intercept at several levels. Process level, agent level and tool call level.

* Process Level
  * An extension to load skills
  * Add keyboard shortcuts
  * Adjust theme
  * Add / commands
  * Add providers
* Agent Level
  * Inject prompt
* Tool calls
  * Show or hide bash scripts
  * Add new tool calls
  * Intercept before and after, think of something like PII detection or raw keys or something or allow the ability to read/write files with allow list or permissions.

## Interactions

* Store history in ~/.bob/history.jsonl which is parsed by the go code and allow resume, by default show first the most recent ones that match the current directory via --resume cli flag

## Built in commands

* /reload - Reload all extensions and skills
* /quit /exit - Leave the program, also ctrl+c should work
* /model - show all the models we have API keys for

## Built in tools that are just extensions shipped

Extensions need a json file that describes the keyword, short description and at what level to load. These are just tools/hooks that the agents can call.

* read - tinygo code to read a file and return the contents
* write - tinygo code to write a file
* bash - tinygo code run a bash script

# Config
Exists in ~/.bob, extensions get ~/.bob/extensions/<name> where they can read/write whatever they want. This should be available through bespoke read/writes that sandbox with the new go fs that doesnt allow .. traversal. Soemthing like ReadConfig(NAME) []byte and WriteConfig(NAME,[]byte)
