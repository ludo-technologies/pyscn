# Catalogue des règles

pyscn propose 33 règles réparties sur 7 catégories. Chaque règle dispose d'une page décrivant ce qu'elle détecte, pourquoi il s'agit d'un problème, un mauvais exemple et comment corriger.

Cliquez sur le nom d'une règle pour ouvrir sa page.

## Code inatteignable

Code mort qui ne peut jamais s'exécuter. Détecté grâce à l'analyse d'accessibilité sur le graphe de flot de contrôle.

| Règle | Sévérité |
| ---- | -------- |
| [`unreachable-after-return`](unreachable-after-return.md) | Critique |
| [`unreachable-after-raise`](unreachable-after-raise.md) | Critique |
| [`unreachable-after-break`](unreachable-after-break.md) | Critique |
| [`unreachable-after-continue`](unreachable-after-continue.md) | Critique |
| [`unreachable-after-infinite-loop`](unreachable-after-infinite-loop.md) | Avertissement |
| [`unreachable-branch`](unreachable-branch.md) | Avertissement |

## Code dupliqué

Fragments de code copiés-collés ou quasi identiques au sein du projet.

| Règle | Sévérité |
| ---- | -------- |
| [`duplicate-code-identical`](duplicate-code-identical.md) | Avertissement |
| [`duplicate-code-renamed`](duplicate-code-renamed.md) | Avertissement |
| [`duplicate-code-modified`](duplicate-code-modified.md) | Info (activation manuelle) |
| [`duplicate-code-semantic`](duplicate-code-semantic.md) | Avertissement |

## Complexité

Fonctions trop ramifiées pour être testées ou raisonnées de manière fiable.

| Règle | Sévérité |
| ---- | -------- |
| [`high-cyclomatic-complexity`](high-cyclomatic-complexity.md) | Selon seuil |

## Conception des classes

Classes qui dépendent de trop d'éléments ou qui assument trop de responsabilités sans rapport.

| Règle | Sévérité |
| ---- | -------- |
| [`high-class-coupling`](high-class-coupling.md) | Selon seuil |
| [`low-class-cohesion`](low-class-cohesion.md) | Selon seuil |

## Injection de dépendances

Schémas de constructeurs et de collaborateurs qui nuisent à la testabilité.

| Règle | Sévérité |
| ---- | -------- |
| [`too-many-constructor-parameters`](too-many-constructor-parameters.md) | Avertissement |
| [`global-state-dependency`](global-state-dependency.md) | Erreur |
| [`module-variable-dependency`](module-variable-dependency.md) | Avertissement |
| [`singleton-pattern-dependency`](singleton-pattern-dependency.md) | Avertissement |
| [`concrete-type-hint-dependency`](concrete-type-hint-dependency.md) | Info |
| [`concrete-instantiation-dependency`](concrete-instantiation-dependency.md) | Avertissement |
| [`service-locator-pattern`](service-locator-pattern.md) | Avertissement |

## Structure des modules

Problèmes du graphe d'imports : cycles, chaînes longues, violations de couches.

| Règle | Sévérité |
| ---- | -------- |
| [`circular-import`](circular-import.md) | Selon la taille du cycle |
| [`deep-import-chain`](deep-import-chain.md) | Info |
| [`layer-violation`](layer-violation.md) | Selon la règle d'architecture |
| [`low-package-cohesion`](low-package-cohesion.md) | Avertissement |
| [`single-responsibility`](single-responsibility.md) | Avertissement / Erreur |

## Données factices

Données fictives livrées par accident en production.

| Règle | Sévérité |
| ---- | -------- |
| [`mock-keyword-in-code`](mock-keyword-in-code.md) | Info / Avertissement |
| [`mock-domain-in-string`](mock-domain-in-string.md) | Avertissement |
| [`mock-email-address`](mock-email-address.md) | Avertissement |
| [`placeholder-phone-number`](placeholder-phone-number.md) | Avertissement |
| [`placeholder-uuid`](placeholder-uuid.md) | Avertissement |
| [`placeholder-comment`](placeholder-comment.md) | Info |
| [`repetitive-string-literal`](repetitive-string-literal.md) | Info |
| [`test-credential-in-code`](test-credential-in-code.md) | Avertissement |

## Sélectionner les règles en ligne de commande

La plupart des utilisateurs exécutent toutes les règles avec `pyscn analyze`. Pour la CI, filtrez par catégorie d'analyseur :

```bash
pyscn check --select deadcode          # uniquement les règles de code inatteignable
pyscn check --select clones            # uniquement les règles de code dupliqué
pyscn check --select complexity        # uniquement high-cyclomatic-complexity
pyscn check --select deps              # circular-import + deep-import-chain + layer-violation
pyscn check --select di                # toutes les règles d'injection de dépendances (activation manuelle)
pyscn check --select mockdata          # toutes les règles de données factices (activation manuelle)
pyscn check --select complexity,deadcode,deps   # combinaison
```

Consultez [`pyscn check`](../cli/check.md) pour la liste complète des options.

## Signification des sévérités

| Sévérité | Intention |
| -------- | --- |
| **Critique** | Presque toujours un bug. À corriger avant la fusion. |
| **Erreur** | Schéma à haut risque. Doit en général faire échouer la CI. |
| **Avertissement** | À examiner. Seuil d'échec par défaut pour `pyscn check`. |
| **Info** | Informatif. N'apparaît que lorsque `min_severity = "info"` ou équivalent. |
| **Selon seuil** | La sévérité dépend d'un seuil numérique (voir les options de la règle). |
