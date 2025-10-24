"""
main.py
The main program used to execute cyclic dependency tests.
"""
from testdata.python.circular.player import create_player

if __name__ == "__main__":
    player = create_player("Hero", 75.0)
    player.update()
