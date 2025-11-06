"""
physics.py
A module for simulating simple gravity and collisions.
"""
from player import Player


def apply_gravity(entity):
    # Apply gravity to any entity
    if isinstance(entity, Player):
        print(f"Applying gravity to {entity.name}")
        entity.position -= 9.8
    else:
        print("apply_gravity: Unsupported entity type")
