"""
physics.py
A module for simulating simple gravity and collisions.
This module will import player.py in reverse to check the Player class.
"""
from player import Player


def apply_gravity(entity):
    # Apply gravity to any entity
    if isinstance(entity, Player):
        print(f"Applying gravity to {entity.name}")
        entity.position -= 9.8
    else:
        print("apply_gravity: Unsupported entity type")


def detect_collision(a, b):
    # Simple Collision Detection Function
    if hasattr(a, "position") and hasattr(b, "position"):
        return abs(a.position - b.position) < 1.0
    return False
