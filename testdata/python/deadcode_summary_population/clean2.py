"""Four functions plus one executable class suite, no dead code."""

FLAGS = ["a", "b"]


class Registry:
    for flag in FLAGS:
        locals()[flag] = flag.upper()

    def lookup(self, flag):
        return getattr(self, flag, None)


def first(value):
    return value


def second(value):
    return value


def third(value):
    return value
