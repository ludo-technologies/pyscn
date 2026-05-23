# Intégration MCP

`pyscn-mcp` est un serveur Model Context Protocol qui expose les analyseurs de pyscn comme outils utilisables par les clients MCP (Claude Code, Cursor, ChatGPT desktop, etc.).

## Outils

| Outil | Équivalent CLI |
| --- | --- |
| `analyze_code` | `pyscn analyze` |
| `check_complexity` | Analyseur de complexité |
| `detect_clones` | Détecteur de clones |
| `check_coupling` | Analyseur CBO |
| `find_dead_code` | Analyseur de code mort |
| `get_health_score` | Score récapitulatif |

Tous les outils acceptent des arguments de chemin et des surcharges de seuils optionnelles. Les résultats sont du JSON structuré.

## Installation

| Méthode | Commande |
| --- | --- |
| uvx (à la demande) | — |
| uv tool | `uv tool install pyscn` |
| pipx | `pipx install pyscn` |
| pip | `pip install pyscn` |

## Configuration des clients

### Claude Code / Claude Desktop

Éditez `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) ou `%APPDATA%/Claude/claude_desktop_config.json` (Windows) :

```json
{
  "mcpServers": {
    "pyscn": {
      "command": "uvx",
      "args": ["pyscn-mcp"]
    }
  }
}
```

Redémarrez l'application.

### Cursor

Settings → Features → Model Context Protocol → Add server :

```json
{
  "pyscn": {
    "command": "uvx",
    "args": ["pyscn-mcp"]
  }
}
```

### Version figée

```json
{
  "mcpServers": {
    "pyscn": {
      "command": "uvx",
      "args": ["pyscn-mcp==0.2.0"]
    }
  }
}
```

### Fichier de configuration personnalisé

```json
{
  "mcpServers": {
    "pyscn": {
      "command": "uvx",
      "args": ["pyscn-mcp"],
      "env": {
        "PYSCN_CONFIG": "/abs/path/to/.pyscn.toml"
      }
    }
  }
}
```

Sans `PYSCN_CONFIG`, le serveur découvre la configuration en remontant depuis le chemin analysé.

### Binaire installé

=== "macOS"

    ```json
    {
      "mcpServers": {
        "pyscn": {
          "command": "/Users/you/Library/Application Support/uv/tools/pyscn/bin/pyscn-mcp"
        }
      }
    }
    ```

=== "Linux"

    ```json
    {
      "mcpServers": {
        "pyscn": {
          "command": "/home/you/.local/share/uv/tools/pyscn/bin/pyscn-mcp"
        }
      }
    }
    ```

=== "Windows"

    ```json
    {
      "mcpServers": {
        "pyscn": {
          "command": "C:/Users/you/AppData/Local/uv/tools/pyscn/bin/pyscn-mcp.exe"
        }
      }
    }
    ```

## Exemple d'invite

> Exécute pyscn sur ce projet et dis-moi quoi corriger en premier.

## Tester

```bash
uvx pyscn-mcp
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | uvx pyscn-mcp
npx @modelcontextprotocol/inspector uvx pyscn-mcp
```

## Modèle de sécurité

- Lecture seule : analyse statique, aucune exécution de code.
- Chemins validés contre la traversée de répertoires.
- Délais d'attente et limites de mémoire par invocation.

Le serveur peut lire tout fichier que le processus appelant peut lire.

## Limitations

- Pas de mode incrémental ; chaque appel relance l'analyse depuis zéro.
- `detect_clones` sur des dépôts de 10 000 fichiers peut prendre plus de 30 secondes.
- Aucun outil d'écriture ; le refactoring s'appuie sur les propres outils d'édition de fichiers de l'assistant.

## Voir aussi

- [Référence CLI](../cli/index.md)
- [Configuration](../configuration/index.md)
