"""Module B - Part of a simple 2-module circular dependency."""

import module_a


class ServiceB:
    """Service B that depends on Service A."""

    def __init__(self):
        self.service_a = None

    def use_service_a(self):
        """Use functionality from Service A."""
        if self.service_a is None:
            self.service_a = module_a.ServiceA()
        return self.service_a.process_request("data from B")

    def process_data(self, data):
        """Process data and potentially use Service A."""
        return f"ServiceB processed: {data}"


def helper_function_b():
    """Helper function in module B."""
    return "Helper B"
