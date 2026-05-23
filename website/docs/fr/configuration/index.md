# Configuration

pyscn lit la configuration depuis `.pyscn.toml` ou la section `[tool.pyscn]` de `pyproject.toml`. Les options en ligne de commande priment sur la configuration ; la configuration prime sur les valeurs par défaut intégrées.

- **[Format du fichier de configuration](format.md)** — Règles de découverte, emplacements des fichiers, ordre de priorité.
- **[Référence](reference.md)** — Chaque clé de configuration, son type, sa valeur par défaut et son effet.
- **[Exemples](examples.md)** — CI stricte, gros codebase, surcharges minimales.

Générez un fichier de démarrage commenté avec :

```bash
pyscn init
```
