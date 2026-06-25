import billing.tax


def total(amount: float) -> float:
    return amount * (1 + billing.tax.RATE)