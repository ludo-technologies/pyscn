# Rapport HTML

Spécification de la sortie HTML générée par `pyscn analyze` (et par défaut lorsque aucun drapeau `--json`/`--yaml`/`--csv` n'est fourni).

## Caractéristiques du fichier

| Propriété | Valeur |
| --- | --- |
| Chemin | `.pyscn/reports/analyze_YYYYMMDD_HHMMSS.html` |
| Encodage | UTF-8 |
| Ressources externes | Aucune (CSS et JS intégrés) |
| Dépendances | Aucune (pas de CDN, pas de polices chargées à distance) |
| Taille | Généralement entre 50 et 500 Ko |

Le fichier est autonome : on peut l'archiver, l'envoyer par e-mail ou le servir depuis n'importe quel hébergeur statique en toute sécurité.

## Structure du document

| Élément | Contenu |
| --- | --- |
| Barre supérieure | Version de pyscn, nom et racine du projet, horodatage de génération, nombre de fichiers, durée. |
| Barre d'onglets | Overview plus un onglet par domaine exécuté, chacun avec un badge de comptage. |
| Overview | Anneau du Health Score et note, un paragraphe de verdict, chiffres de taille, cartes de score par dimension, suggestions prioritaires, fichiers sensibles, histogramme de complexité, et cartes de synthèse pour la duplication, les classes et la structure. |
| Onglets de détail | Functions, Duplication, Classes, Architecture. |
| Pied de page | Lien vers le dépôt pyscn et chaîne de version. |

Les cartes de score, cartes de synthèse et onglets n'apparaissent que pour les analyseurs exécutés. Architecture apparaît si l'analyse des dépendances, de l'architecture ou des communautés a été exécutée ; les règles de couches nécessitent des couches `[architecture]` configurées.

## Overview

| Bloc | Contenu |
| --- | --- |
| Verdict | Health Score (0–100) dessiné en anneau, badge de note (A–F), un titre selon la note, et une phrase nommant les dimensions saines et les deux ou trois plus faibles avec leurs chiffres clés. |
| Score breakdown | Une carte par dimension activée (Complexity, Dead code, Duplication, Coupling, Cohesion, Dependencies, Architecture, Communities) avec son score 0–100, une barre colorée par bande et deux chiffres complémentaires. Les cartes renvoient vers l'onglet de détail correspondant. |
| Fix first | Les cinq suggestions les plus prioritaires avec sévérité, effort, emplacement et justification. |
| Hotspot files | Jusqu'à huit modules classés par fonctions à haut risque, puis complexité maximale, puis code mort, puis fragments clonés. |
| Complexity distribution | Histogramme de la complexité des fonctions découpé selon les seuils de risque configurés, avec médiane, imbrication la plus profonde et fonction la plus longue. |
| Duplication / Classes / Structure | Cartes de synthèse compactes renvoyant vers leurs onglets de détail. |

## Onglets de détail

| Onglet | Contenu |
| --- | --- |
| Functions | Bandeau de métriques de complexité, fonctions les plus complexes (top 20), fonctions les plus longues, constats de code mort (top 20), et deux tableaux triables repliés : tous les modules et les agrégats de complexité par répertoire. |
| Duplication | Bandeau de statistiques de clones, groupes de clones (top 10) avec fragments et aperçus de code optionnels, ou paires de clones si aucun groupe ne s'est formé. |
| Classes | Bandeaux de couplage (CBO) et de cohésion (LCOM4) avec les classes les plus couplées et les moins cohésives (top 15 chacune). |
| Architecture | Métriques de dépendances entre modules, zones de la séquence principale, dépendances circulaires, chaînes les plus longues, violations des règles de couches, et détection de communautés avec le graphe de macro-architecture. |

## JavaScript

Des scripts inline changent d'onglet (l'onglet est reflété dans le hash de l'URL, un lien peut donc ouvrir un onglet précis) et font passer les tableaux de modules et de répertoires par un trieur commun. Aucune requête réseau n'est effectuée.

Les agrégats de modules utilisent la population complète de l'analyseur avant les filtres d'affichage `min_complexity`, `report_unchanged` ou `min_severity`. La complexité par répertoire utilise la population de fonctions rapportées après les filtres de complexité, de sorte que ses comptes et moyennes concordent avec l'onglet Functions.

## CSS

Les styles sont inlinés depuis `service/templates/analyze/report.css` et utilisent des propriétés CSS personnalisées. Principaux jetons :

| Variable | Rôle sémantique |
| --- | --- |
| `--good` / `--warn` / `--bad` | Bandes de score (75 et plus, 60–74, moins de 60), niveaux de risque, sévérités. |
| `--accent` | Navigation, liens, barres de graphique neutres. |
| `--ink` / `--ink-2` / `--muted` | Texte principal, secondaire et légendes. |
| `--surface` / `--page` / `--line` | Cartes, fond de page et filets. |

Le mode sombre suit la media query `prefers-color-scheme`, et un attribut `data-theme="light"` ou `data-theme="dark"` sur l'élément racine le force. Aucun bouton de bascule n'est rendu.

## Ouverture automatique

Le rapport s'ouvre dans le navigateur par défaut lorsque **toutes** les conditions suivantes sont remplies :

- Le format est HTML.
- Stdin est un TTY.
- Les variables d'environnement `SSH_TTY` et `SSH_CONNECTION` ne sont pas définies.
- La variable d'environnement `CI` n'est pas définie.
- `--no-open` n'est pas passé.

Mécanisme d'ouverture : `open` sur macOS, `xdg-open` (ou `gnome-open` / `kde-open`) sous Linux, `cmd /c start` sous Windows.

Forme de l'URL de fichier : `file:///{chemin-absolu-vers-le-rapport}`.

Le chemin du rapport est toujours affiché sur stderr, indépendamment de l'ouverture automatique.

## Désactiver l'ouverture automatique

```bash
pyscn analyze --no-open .
```

Ou exportez `CI=true` dans l'environnement.

## Correspondance des badges de note

| Note | Score | Couleur du badge |
| ---- | ----- | --- |
| A | 90–100 | Vert (`--good`) |
| B | 75–89  | Vert (`--good`) |
| C | 60–74  | Ambre (`--warn`) |
| D | 45–59  | Rouge (`--bad`) |
| F | 0–44   | Rouge (`--bad`) |

## Références croisées

- [Score de santé](health-score.md) — formule du score global.
- [Schémas](schemas.md) — alternatives lisibles par une machine.
- [Formats de sortie](index.md) — tous les formats de sortie et le contrat de stabilité.
