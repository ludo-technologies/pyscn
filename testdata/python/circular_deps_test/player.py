"""
player.py
A module representing the game character.
This module depends on apply_gravity in physics.py.
"""
from physics import apply_gravity


class Player:
    def __init__(self, name: str, mass: float):
        self.name = name
        self.mass = mass
        self.position = 0

    def update(self):
        # Apply Gravity to Player
        # Gravity for Player Applications
        apply_gravity(self)
        print(f"{self.name} updated position to {self.position}")


def create_player(name: str, mass: float) -> "Player":
    return Player(name, mass)
