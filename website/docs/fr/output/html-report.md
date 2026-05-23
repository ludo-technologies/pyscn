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
| En-tête | Nom du projet, horodatage de génération, version de pyscn, durée. |
| Carte du score global | Score de santé (0–100), badge de note (A–F). |
| Cartes de score par catégorie | Une par analyseur activé avec son score sur 0–100. |
| Onglets | Résumé, Complexité, Code mort, Clones, Couplage, Cohésion, Dépendances, Architecture. |
| Pied de page | Lien vers le dépôt pyscn et chaîne de version. |

Les cartes de score par catégorie et les onglets n'apparaissent que pour les analyseurs qui ont été exécutés. L'onglet Architecture n'apparaît que si des couches `[architecture]` sont configurées.

## Onglets

| Onglet | Contenu |
| --- | --- |
| Résumé | Chiffres globaux et note. |
| Complexité | Tableau triable des fonctions avec complexité McCabe / cognitive, profondeur d'imbrication, risque. |
| Code mort | Constats regroupés par sévérité avec fichier:ligne et motif. |
| Clones | Groupes de clones avec similarité et type de clone. |
| Couplage | Classes par CBO avec ventilation par type de dépendance. |
| Cohésion | Classes par LCOM4 avec regroupement des méthodes. |
| Dépendances | Graphe de modules, métriques Ca/Ce/I/A/D, cycles. |
| Architecture | Violations des règles entre couches. |

## JavaScript

Une seule fonction intégrée, `showTab(id)`, permet de basculer entre les onglets. Aucun autre script ne s'exécute. Aucune requête réseau.

## CSS

Les styles utilisent des propriétés personnalisées CSS portant ces noms :

| Variable | Rôle sémantique |
| --- | --- |
| `--color-success` | Constats à faible risque, note A. |
| `--color-warning` | Constats à risque moyen, notes B/C. |
| `--color-danger` | Constats à risque élevé, notes D/F. |
| `--color-text` | Texte courant. |
| `--color-muted` | Texte secondaire. |

Le mode sombre suit la requête média `prefers-color-scheme` ; aucun interrupteur n'est fourni.

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

| Note | Score | Couleur de fond du badge |
| ---- | ----- | --- |
| A | 90–100 | Vert (`--color-success`) |
| B | 75–89  | Teinté vert |
| C | 60–74  | Orange (`--color-warning`) |
| D | 45–59  | Teinté rouge (`--color-danger`) |
| F | 0–44   | Rouge (`--color-danger`) |

## Références croisées

- [Score de santé](health-score.md) — formule du score global.
- [Schémas](schemas.md) — alternatives lisibles par une machine.
- [Formats de sortie](index.md) — tous les formats de sortie et le contrat de stabilité.
