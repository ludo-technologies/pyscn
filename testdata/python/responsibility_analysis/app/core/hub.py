"""Cross-cutting hub that intentionally mixes app concerns."""

from ..api.views import render
from ..auth.policy import authorize
from ..billing.invoice import charge
from ..db.repo import save
from ..reporting.export import export


def process(user, payload):
    authorize(user)
    result = render(payload)
    save(result)
    charge(user, result)
    return export(result)
