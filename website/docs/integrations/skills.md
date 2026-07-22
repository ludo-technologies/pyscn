# Agent Skills

pyscn ships 4 Agent Skills that teach coding agents when and how to run each analysis — no MCP server required.

## Installation

```bash
uvx add-skills ludo-technologies/pyscn
```

Installs the Skills into your project. Works with Claude Code, Cursor, Codex, Gemini CLI, and [many other agents](https://github.com/ludo-technologies/add-skills) (add `--agent cursor` etc. to target one, `--global` for all projects).

## Skills

| Skill | Use it when |
| --- | --- |
| `health-check` | "How healthy is this code?", a quality overview, a before/after comparison |
| `refactoring` | Finding refactoring targets — duplicate code, complexity hotspots, dead code |
| `architecture-review` | Module structure, coupling, circular dependencies, which files to review together |
| `cli-analysis` | CI/CD quality gates, shareable reports, project configuration |

Each Skill runs `uvx pyscn@latest <command>` under the hood — no install required to try it.

## Example prompts

> Analyze the code quality of the app/ directory

> Find duplicate code and help me refactor it

> Show me complex code and help me simplify it

> Check for circular dependencies before I merge this

## Claude Code plugin

Installs the MCP server and the Skills together:

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

## Skills vs. MCP

Skills teach an agent when to reach for pyscn and which CLI command to run; they need no server and work with any agent that supports the Skill format. The [MCP server](mcp.md) instead exposes the same analyses as structured tool calls — use it when you want typed JSON results wired directly into the client rather than shell output. The two are complementary and can be installed together via the Claude Code plugin above.

## See also

- [MCP Integration](mcp.md)
- [CLI Reference](../cli/index.md)
- [add-skills](https://github.com/ludo-technologies/add-skills)
