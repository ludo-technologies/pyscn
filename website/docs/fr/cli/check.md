# `pyscn check`

Garde-fou qualité pour les pipelines CI/CD. Écrit les constatations au format linter sur **stderr** et sort avec un code non nul si un problème dépasse un seuil.

```text
pyscn check [flags] [paths...]
```

Les chemins sont par défaut le répertoire courant.

## Ce qu'elle fait

`check` est le compagnon CI de [`analyze`](analyze.md) :

- **Les constatations vont sur stderr** au format linter (`file:line:col: message`).
- **Sortie 0** en cas de succès, **sortie 1** si des problèmes sont détectés, **sortie 2** si l'analyse n'a pas pu aboutir.
- **Valeurs par défaut strictes** — toute fonction de complexité supérieure à 10 échoue ; toute dépendance circulaire échoue (lorsque `--select deps` est défini) ; tout fichier impossible à analyser syntaxiquement échoue.
- **Rapide** — n'exécute que les analyses sélectionnées ; pas de génération de rapport.

!!! note "Portée des erreurs d'analyse syntaxique"
    Toutes les analyses sélectionnées utilisent le même registre de découverte du projet. Une erreur d'analyse syntaxique fait échouer la vérification, y compris avec `pyscn check --select deps`, sauf si `--allow-parse-errors` est défini. Une erreur de lecture reste toujours bloquante.

## Options

### Sélection des analyses

| Option | Description |
| --- | --- |
| `-s, --select <list>` | N'exécute que les analyses listées. Valeurs : `complexity`, `deadcode`, `clones`, `deps` (alias `circular`), `mockdata`, `di`. |
| `--skip-clones`       | N'exécute pas la détection de clones. |

Par défaut (sans `--select`) : exécute `complexity`, `deadcode`, **et `clones`**. `deps`, `mockdata` et `di` sont en opt-in via `--select`. Passez `--skip-clones` pour ignorer la détection de clones sans passer à `--select`.

### Surcharges de seuils

| Option | Défaut | Description |
| --- | --- | --- |
| `--max-complexity <N>`   | `10` | Échoue si une fonction dépasse cette complexité cyclomatique. |
| `--max-cycles <N>`       | `0`  | Nombre maximal de cycles de dépendance circulaire avant échec. |
| `--allow-dead-code`      | off  | Traite le code mort comme un simple avertissement ; ne fait pas échouer la vérification. |
| `--allow-circular-deps`  | off  | Traite les cycles comme de simples avertissements ; ne fait pas échouer la vérification. |
| `--allow-parse-errors`   | off  | Traite les erreurs d'analyse syntaxique comme de simples avertissements. Les erreurs de lecture restent bloquantes. |

Par défaut, un fichier impossible à analyser syntaxiquement fait échouer la vérification. Un tel fichier est exclu de toutes les analyses : il ne produit aucune constatation et franchirait donc tous les seuils — une erreur de syntaxe dans vos sources serait rapportée comme une exécution propre. L'option ne masque ni les fichiers absents, ni les erreurs d'autorisation, ni les autres erreurs de lecture.

### Sortie

| Option | Description |
| --- | --- |
| `-q, --quiet`          | Supprime la sortie sauf si des problèmes sont détectés. |
| `-c, --config <path>`  | Charge la configuration depuis un fichier spécifique. |
| `-v, --verbose`        | Affiche la progression détaillée. |

## Codes de sortie

| Code | Signification |
| --- | --- |
| `0` | Toutes les vérifications ont réussi. |
| `1` | Un ou plusieurs seuils de qualité ont été dépassés. |
| `2` | L'analyse n'a pas pu aboutir sur les cibles demandées — entrée invalide, fichiers manquants ou impossibles à analyser. |

La sortie `1` est un verdict sur votre code ; la sortie `2` signifie que le verdict lui-même est incomplet et ne doit pas être interprété comme un succès.

## Exemples

```bash
# Garde-fou CI standard (exécute complexity, deadcode, clones)
pyscn check .

# Garde-fou plus rapide : ignore la détection de clones
pyscn check --skip-clones .

# Complexité uniquement, avec un seuil plus élevé pour du code hérité
pyscn check --select complexity --max-complexity 15 src/

# Vérifier les imports circulaires
pyscn check --select deps src/

# Tolérer le code mort existant pendant le nettoyage
pyscn check --allow-dead-code src/

# Détecter les anti-patterns DI (opt-in)
pyscn check --select di src/

# Mode silencieux — idéal pour les journaux CI
pyscn check --quiet .
```

## Relation avec `analyze`

`check` utilise les mêmes analyseurs et le même fichier de configuration qu'`analyze`. Les différences :

| Aspect | `analyze` | `check` |
| --- | --- | --- |
| Sortie | Fichier de rapport (HTML/JSON/YAML/CSV) | stderr au format linter |
| Sortie en cas de problèmes | Toujours `0` (sauf erreur) | Sortie `1` si un problème dépasse un seuil |
| Détection de clones | Activée par défaut | Activée par défaut (désactiver avec `--skip-clones`) |
| Analyse des dépendances | Activée par défaut | Désactivée par défaut (opt-in via `--select deps`) |
| Vitesse | Plus lent (tous les analyseurs, génération de rapport) | Rapide (uniquement les analyses sélectionnées, sans rapport) |
| Cas d'usage | Revue interactive | Garde-fou qualité en CI |

Utilisez les deux : `analyze` pour comprendre les problèmes, `check` pour prévenir les régressions.

## Voir aussi

- [Intégration CI/CD](../integrations/ci-cd.md) — exemples GitHub Actions / pre-commit / GitLab.
- [`pyscn analyze`](analyze.md) — Analyse complète avec rapports.
