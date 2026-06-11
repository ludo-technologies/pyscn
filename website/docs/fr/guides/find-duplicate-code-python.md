---
title: Comment détecter le code dupliqué en Python
description: Un guide pratique pour détecter le code dupliqué en Python. Découvrez ce que sont les clones de code (Types 1 à 4), quels outils les détectent et comment automatiser la détection en CI.
---

# Comment détecter le code dupliqué en Python

Le code dupliqué pose de vrais problèmes. Un bug corrigé dans une copie survit dans les autres. Chaque modification doit être répercutée à plusieurs endroits. Les relecteurs perdent du temps à lire une logique qu'ils ont déjà vue. Et maintenant que les assistants IA écrivent une grande partie de notre code, des blocs quasi identiques apparaissent dans les bases de code plus vite que les humains ne les copiaient-collaient.

Ce guide explique ce qui constitue du code « dupliqué », quels outils permettent de le détecter en Python, et comment lancer la détection en local et en CI.

## Réponse rapide

Pour analyser un projet immédiatement, lancez :

```bash
uvx pyscn@latest analyze --select clones .
```

Cette commande exécute la détection de clones de [pyscn](https://github.com/ludo-technologies/pyscn) sans rien installer. Elle ouvre un rapport HTML listant tous les groupes de code dupliqué, avec les scores de similarité et les emplacements dans les fichiers.

## Qu'est-ce que du code « dupliqué » ? Les quatre types de clones

La plupart des développeurs imaginent le code dupliqué comme un simple copier-coller. Dans la recherche, les duplicats sont appelés *clones de code* et se répartissent en quatre types. La distinction est importante, car la plupart des outils ne détectent que les un ou deux premiers.

Voici la même fonction déclinée en quatre variantes.

**Type 1 : code identique.** Un copier-coller où seuls l'espacement et les commentaires diffèrent :

```python
def calculate_order_total(items, discount_rate):
    subtotal = 0.0
    for item in items:
        price = item["price"]
        quantity = item["quantity"]
        if quantity <= 0:
            continue
        subtotal += price * quantity
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    tax = subtotal * 0.1
    total = subtotal + tax
    return round(total, 2)
```

Si ce bloc est collé dans un autre fichier, peut-être avec un commentaire ajouté, c'est un clone de Type 1. C'est le type le plus facile à détecter et à corriger : il suffit d'extraire une fonction commune. ([règle : duplicate-code-identical](../rules/duplicate-code-identical.md))

**Type 2 : identifiants renommés.** La structure est intacte, seuls les noms ont changé :

```python
def compute_cart_amount(products, rebate):
    amount = 0.0
    for product in products:
        cost = product["price"]
        count = product["quantity"]
        if count <= 0:
            continue
        amount += cost * count
    if rebate > 0:
        amount = amount * (1 - rebate)
    levy = amount * 0.1
    result = amount + levy
    return round(result, 2)
```

Les outils basés sur les lignes ratent ce cas, car aucune ligne n'est textuellement identique. Mais si l'on compare les arbres syntaxiques avec les noms normalisés, les deux fonctions ont exactement la même forme. ([règle : duplicate-code-renamed](../rules/duplicate-code-renamed.md))

**Type 3 : copies modifiées.** Quelqu'un a copié la fonction, puis ajouté ou supprimé quelques instructions :

```python
def calculate_quote_total(items, discount_rate, shipping=0.0):
    subtotal = 0.0
    for item in items:
        price = item["price"]
        quantity = item["quantity"]
        if quantity <= 0:
            continue
        subtotal += price * quantity
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    subtotal += shipping        # <- nouveau
    tax = subtotal * 0.1
    total = subtotal + tax
    return round(total, 2)
```

C'est le type de clone le plus courant dans les vraies bases de code. Quelqu'un a copié une fonction, l'a adaptée à un nouveau cas et est passé à autre chose. La détecter implique de mesurer l'écart entre deux arbres (tree edit distance), pas seulement de vérifier s'ils correspondent. ([règle : duplicate-code-modified](../rules/duplicate-code-modified.md))

**Type 4 : même comportement, implémentation différente.** Le code a été réécrit de zéro mais calcule la même chose :

```python
def total_for_order(items, discount_rate):
    valid_items = []
    for item in items:
        if item["quantity"] > 0:
            valid_items.append(item)
    subtotal = sum(
        item["price"] * item["quantity"]
        for item in valid_items
    )
    if discount_rate > 0:
        subtotal = subtotal * (1 - discount_rate)
    total_with_tax = subtotal * 1.1
    return round(total_with_tax, 2)
```

Aucune correspondance textuelle ou structurelle ne reliera ce code à l'original. La comparaison de la structure du flux de contrôle le peut. ([règle : duplicate-code-semantic](../rules/duplicate-code-semantic.md))

## Outils de détection du code dupliqué en Python

Voici les principales options et ce que chacune peut détecter :

| Outil | Détecte | Remarques |
| --- | --- | --- |
| [pylint](https://pylint.readthedocs.io/) (`R0801`) | Type 1 | Vérification de similarité basée sur les lignes, inclus dans pylint. Détecte le copier-coller. Les renommages le trompent. |
| [jscpd](https://github.com/kucherenko/jscpd) | Type 1, Type 2 partiel | Basé sur les tokens, supporte plus de 150 langages. Adapté si un seul détecteur doit couvrir un dépôt multilangage. |
| [SonarQube](https://www.sonarsource.com/products/sonarqube/) | Type 1, Type 2 partiel | Plateforme complète avec tableaux de bord et historique. Plus lourd à configurer et à héberger. |
| [PMD CPD](https://pmd.github.io/pmd/pmd_userdocs_cpd.html) | Type 1, Type 2 | Le détecteur de copier-coller classique. Nécessite une JVM. |
| [pyscn](https://github.com/ludo-technologies/pyscn) | Types 1 à 4 | Spécifique à Python. Utilise le hachage AST pour les Types 1-2, la tree edit distance (APTED) pour le Type 3 et la comparaison de flux de contrôle pour le Type 4. |

Un point de confusion fréquent : **[ruff](https://docs.astral.sh/ruff/) ne détecte pas le code dupliqué.** Ruff est un linter et un formateur. Il vérifie la façon dont les lignes et instructions individuelles sont écrites, ce qui est un travail différent de la comparaison de fonctions entre elles à travers les fichiers. Les deux types d'outils se complètent plutôt qu'ils ne se concurrencent.

## Tutoriel : détecter les clones avec pyscn

Reprenez les quatre variantes ci-dessus et répartissez-les dans deux fichiers, `orders.py` et `invoices.py`, avec l'original collé dans les deux. Ensuite, lancez :

```bash
uvx pyscn@latest analyze --select clones .
```

pyscn analyse tous les fichiers Python, extrait les fragments de code et les compare par paires. Sur les grandes bases de code, il utilise [LSH](https://en.wikipedia.org/wiki/Locality-sensitive_hashing) pour rester rapide, à plus de 100 000 lignes par seconde. Le résumé s'affiche dans le terminal :

```text
📊 Analysis Summary:
Health Score: 80/100 (Grade: B)

📈 Detailed Scores:
  Duplication:      0/100 ❌  (10.0% duplication, 1 groups)
```

Le rapport HTML regroupe les cinq fragments en un seul groupe de clones, avec chaque paire classifiée et scorée :

| Paire | Classifié comme | Similarité |
| --- | --- | --- |
| copie exacte entre les deux fichiers | Type 1 | 1,00 |
| original vs. copie modifiée | Type 2 | 0,85 |
| original vs. version réécrite | Type 4 | 0,94 |

Notez la dernière ligne. La version réécrite `total_for_order` est la variante qu'aucun outil basé sur le texte ne peut relier à l'original. pyscn la détecte avec une similarité de 0,94 à partir de sa structure de flux de contrôle.

### Ajuster le seuil

Le flag `--clone-threshold` (valeur par défaut `0.65`) définit la similarité minimale pour qu'une paire soit signalée :

```bash
pyscn analyze --select clones --clone-threshold 0.8 .   # plus strict : moins de correspondances, plus proches
```

Pour des réglages permanents, créez un fichier `.pyscn.toml` (ou utilisez `[tool.pyscn]` dans `pyproject.toml`) :

```toml
[clones]
similarity_threshold = 0.8
min_lines = 15        # ignore fragments smaller than this
```

Les fonctions très courtes sont ignorées par défaut (`min_lines`). En dessous d'une certaine taille, la similarité perd de son sens : chaque getter de deux lignes ressemble à tous les autres. Consultez la [référence de configuration](../configuration/reference.md#clones) pour toutes les options, y compris l'activation et la désactivation des types de clones individuels.

## Automatiser la détection en CI

`pyscn check` est la version CI de `analyze`. Il ne produit pas de rapport, seulement un code de sortie succès/échec :

```bash
pyscn check --select clones .
```

Comme étape GitHub Actions :

```yaml
- uses: actions/setup-python@v5
  with:
    python-version: "3.12"
- run: pipx run pyscn check --select clones .
```

Le job échoue lorsque la duplication dépasse vos seuils. C'est précisément l'intérêt : une détection qu'il faut penser à lancer est une détection qui finit par ne plus avoir lieu. Consultez [Intégration CI/CD](../integrations/ci-cd.md) pour des workflows complets, et [Pyscn Bot](https://github.com/marketplace/pyscn-bot) pour recevoir des analyses automatiques sur les pull requests.

## Que faire des résultats

Tous les clones n'ont pas besoin d'être supprimés. Un ordre d'attaque utile :

1. **Les clones de Type 1 et Type 2 dans le code de production.** Extrayez une fonction commune. La correction est mécanique et peu risquée, précisément parce que les copies sont quasi identiques.
2. **Les clones de Type 3.** Examinez ce qui diffère entre les copies. Si les différences sont des données, extrayez une fonction qui prend des paramètres. Si les différences sont comportementales, les copies divergent peut-être intentionnellement. Parfois, deux points d'appel ont genuinement besoin d'évoluer séparément, et les fusionner couplerait des éléments qui devraient rester indépendants.
3. **Les clones de Type 4.** Traitez-les comme des signaux plutôt que comme des actions à mener. Deux implémentations indépendantes de la même logique signifient souvent que deux personnes ignoraient les travaux de l'autre. Choisissez l'une des deux, ou documentez pourquoi les deux coexistent.
4. **Les clones dans le code de test.** Soyez plus tolérant ici. Les tests privilégient l'explicite au détriment du DRY, et une certaine répétition qui rend chaque test lisible en lui-même en vaut généralement la peine.

En pratique, il vaut mieux fixer un seuil strict pour que le rapport reste court, corriger le groupe en tête de liste et relancer l'analyse. Tenter d'éliminer un rapport de 40 groupes en un seul grand refactoring aboutit rarement bien.

## FAQ

**Est-ce que ruff détecte le code dupliqué ?**
Non. Ruff est un linter et un formateur et ne possède pas de règles de détection de clones. Trouver des duplicats nécessite de comparer des fragments de code entre eux à travers les fichiers, ce qui dépasse le périmètre d'un linter. Utilisez ruff pour les vérifications de style et de correction, et un détecteur de clones pour la duplication. Les deux fonctionnent bien ensemble.

**Quel niveau de duplication est acceptable ?**
Il n'existe pas de chiffre universel. À titre indicatif, moins de 5 % de lignes dupliquées est typique d'une base de code bien entretenue, et tout ce qui dépasse 15 % signifie généralement un développement systématique par copier-coller. La tendance importe plus que le chiffre lui-même. Une duplication qui croît release après release est le vrai signal d'alarme.

**Puis-je détecter des duplicats sur plusieurs dépôts ?**
Oui. Pointez l'analyseur vers un répertoire contenant les deux checkouts : `pyscn analyze --select clones repo-a/ repo-b/`. Les fragments sont comparés sur tout ce qui est dans le périmètre, donc les clones inter-dépôts apparaissent comme n'importe quelle autre paire.

**Pourquoi mon fragment dupliqué n'est-il pas signalé ?**
Il est très probablement inférieur à la taille minimale de fragment (`min_lines` / `min_nodes` dans la configuration). Les détecteurs ignorent intentionnellement les fragments trop petits. À cinq lignes, la moitié de la base de code ressemble à l'autre moitié. Réduisez les limites dans `.pyscn.toml` si vous souhaitez que les fragments courts soient également comparés.

---

*Suite : parcourez le [catalogue des règles de code dupliqué](../rules/index.md) pour voir comment chaque type de clone est scoré, ou la [documentation du score de santé](../output/health-score.md) pour comprendre comment la duplication influe sur la note de votre projet.*
