import inventory.warehouse


def available(units: int) -> bool:
    return units <= inventory.warehouse.CAPACITY