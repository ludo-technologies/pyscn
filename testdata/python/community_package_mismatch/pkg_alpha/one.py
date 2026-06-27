import pkg_alpha.two
import pkg_beta.one


def run() -> str:
    return pkg_alpha.two.VALUE + pkg_beta.one.label()