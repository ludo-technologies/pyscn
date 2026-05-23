# Formats de sortie

pyscn écrit les résultats d'analyse dans des fichiers situés dans un répertoire de sortie. Tous les formats partagent une sémantique de champ stable d'une version corrective à l'autre.

## Répertoire de sortie

Par défaut : `.pyscn/reports/` sous le répertoire de travail courant.

Configurable via `.pyscn.toml` :

```toml
[output]
directory = "build/reports"
```

## Convention de nommage

```
{command}_YYYYMMDD_HHMMSS.{ext}
```

`{command}` vaut `analyze` (la seule commande pyscn qui écrit des rapports). L'horodatage est en heure locale. Les fichiers existants ne sont jamais écrasés.

## Formats pris en charge

| Format | Extension | Drapeau           | Spécification                    |
| ------ | --------- | ----------------- | -------------------------------- |
| text   | —         | (terminal)        | lisible par un humain, non stable |
| json   | `.json`   | `--json`          | [schemas.md](schemas.md)         |
| yaml   | `.yaml`   | `--yaml`          | [schemas.md](schemas.md)         |
| csv    | `.csv`    | `--csv`           | [schemas.md](schemas.md)         |
| html   | `.html`   | `--html` (défaut) | [html-report.md](html-report.md) |

Le format `text` est destiné à l'affichage dans le terminal et ne fait l'objet d'aucun contrat de stabilité ; sa présentation peut changer entre n'importe quelles versions.

## Contrat de stabilité

Entre versions correctives et mineures d'une même version majeure :

- **Stable** : noms de champs, types et signification sémantique dans `json`, `yaml` et `csv`.
- **Peut changer** : ordre des éléments dans les tableaux, ajout de nouveaux champs, ajout de nouvelles sections de premier niveau, modifications cosmétiques de `text` et `html`.
- **Changements incompatibles** : limités aux changements de version majeure (suppression ou renommage de champs, modification du type d'un champ).

Les intégrations tierces doivent ignorer les champs inconnus et ne pas dépendre de l'ordre des champs au sein d'un objet.
