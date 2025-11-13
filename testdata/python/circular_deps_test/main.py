"""
main.py
The main program used to execute cyclic dependency tests.
"""
from player import Player

if __name__ == "__main__":
    player = Player("Hero", 75.0)
    player.update()
