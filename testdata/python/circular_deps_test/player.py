"""
player.py
A module representing the game character.
"""
from physics import apply_gravity


class Player:
    def __init__(self, name: str, mass: float):
        self.name = name
        self.mass = mass
        self.position = 0

    def update(self):
        # Apply gravity to player
        apply_gravity(self)
        print(f"{self.name} updated position to {self.position}")
